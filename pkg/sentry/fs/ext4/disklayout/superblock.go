// Copyright 2019 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package disklayout

import (
	"time"
)

// SuperBlock should be implemented by structs representing ext4 superblock.
// The superblock holds a lot of information about the enclosing filesystem.
// This interface aims to provide access methods to important information held
// by the superblock. It does NOT expose all fields of the superblock, only the
// ones necessary. This can be expanded when need be.
//
// Location and replication:
//     - The superblock is located at offset 1024 in block group 0.
//     - Redundant copies of the superblock and group descriptors are kept in
//       all groups if sparse_super feature flag is NOT set. If it is set, the
//       replicas only exist in groups whose group number is either 0 or a
//       power of 3, 5, or 7.
//     - There is also a sparse superblock feature v2 in which there are just
//       two replicas saved in block groups pointed by the s_backup_bgs field.
//
// Replicas should eventually be updated if the superblock is updated.
//
// See https://www.kernel.org/doc/html/latest/filesystems/ext4/globals.html#super-block.
type SuperBlock interface {
	// InodesCount returns the total number of inodes in this filesystem.
	InodesCount() uint32

	// BlocksCount returns the total number of data blocks in this filesystem.
	BlocksCount() uint64

	// FreeBlocksCount returns the number of free blocks in this filesystem.
	FreeBlocksCount() uint64

	// FreeInodesCount returns the number of free inodes in this filesystem.
	FreeInodesCount() uint32

	// MountCount returns the number of mounts since the last fsck.
	MountCount() uint16

	// MaxMountCount returns the number of mounts allowed beyond which a fsck is
	// needed.
	MaxMountCount() uint16

	// FirstDataBlock returns the absolute block number of the first data block,
	// which contains the super block itself.
	//
	// If the filesystem has 1kb data blocks then this should return 1. For all
	// other configurations, this typically returns 0.
	//
	// The first block group descriptor is in (FirstDataBlock() + 1)th block.
	FirstDataBlock() uint32

	// FirstInode returns the first non-reserved inode number.
	FirstInode() uint32

	// BlockSize returns the size of one data block in this filesystem.
	// This can be calculated by 2^(10 + sb.s_log_block_size).
	BlockSize() uint64

	// BlocksPerGroup returns the number of data blocks in a block group.
	BlocksPerGroup() uint32

	// ClusterSize returns block cluster size (set during mkfs time by admin).
	//
	// Cluster size is:
	//     - (2^sb.s_log_cluster_size) * BlockSize()   if bigalloc is enabled.
	//     - BlockSize()                               otherwise.
	// Cluster size holds no meaning when bigalloc is not set and hence
	// sb.s_log_cluster_size must equal sb.s_log_block_size in that case.
	ClusterSize() uint64

	// ClustersPerGroup returns:
	//     - number of clusters per group        if bigalloc is enabled.
	//     - BlocksPerGroup()                    otherwise.
	ClustersPerGroup() uint32

	// InodeSize returns the size of the inode struct in bytes.
	InodeSize() uint16

	// InodesPerGroup returns the number of inodes in a block group.
	InodesPerGroup() uint32

	// BgDescSize returns the size of the block group descriptor struct.
	BgDescSize() uint16

	// CompatibleFeatures returns the CompatFeatures struct which holds all the
	// compatible features this fs supports.
	CompatibleFeatures() CompatFeatures

	// IncompatibleFeatures returns the CompatFeatures struct which holds all the
	// incompatible features this fs supports.
	IncompatibleFeatures() IncompatFeatures

	// ReadOnlyCompatibleFeatures returns the CompatFeatures struct which holds all the
	// readonly compatible features this fs supports.
	ReadOnlyCompatibleFeatures() RoCompatFeatures

	// MountTime returns the last mount time.
	MountTime() time.Time

	// WriteTime returns the last write time.
	WriteTime() time.Time

	// CreationTime returns the time when mkfs was run to create this fs.
	CreationTime() time.Time

	// Magic() returns the magic signature which must be 0xef53.
	Magic() uint16

	// State returns the superblock state.
	State() SbState

	// ErrorPolicy returns the superblock error policy which dictates the
	// behaviour when detecting errors.
	ErrorPolicy() SbErrorPolicy

	// CreatorOS returns the OSCode of the operating system which initially
	// created this filesystem.
	CreatorOS() OSCode

	// GroupNumber returns the block group number of this superblock.
	// Can be non-zero if this is a redundant copy.
	GroupNumber() uint16

	// UUID returns the 128-bit UUID for this volume. This is used for
	// checksumming in most ext4 objects.
	UUID() [16]byte

	// Label returns the volume label. Max len is 16.
	Label() string
}

// Superblock compatible features.
const (
	// Directory preallocation.
	SbDirPrealloc = 0x1

	// Imagic Inode. Unclear what this does.
	SbImagicInodes = 0x2

	// Has a journal. jbd2 should only work with this being set.
	SbHasJournal = 0x4

	// Fs supports extended attributes.
	SbExtAttr = 0x8

	// Has reserved GDT blocks (right after group descriptors) for fs expansion.
	SbResizeInode = 0x10

	// Has directory indices.
	SbDirIndex = 0x20

	// Lazy block group. Unclear what this does.
	SbLazyBg = 0x40

	// Exclude Inode. Not used.
	SbExcludeInode = 0x80

	// Exclude bitmap. Not used.
	SbExcludeBitmap = 0x100

	// Sparse superblock version 2.
	SbSparseV2 = 0x200
)

// CompatFeatures represents a superblock's compatible feature set. If the
// kernel does not understand any of these feature, it can still read/write
// to this fs.
type CompatFeatures struct {
	DirPrealloc   bool
	ImagicInodes  bool
	HasJournal    bool
	ExtAttr       bool
	ResizeInode   bool
	DirIndex      bool
	LazyBg        bool
	ExcludeInode  bool
	ExcludeBitmap bool
	SparseV2      bool
}

// ToInt converts superblock compatible features back to its 32-bit rep.
func (f CompatFeatures) ToInt() uint32 {
	var res uint32

	if f.DirPrealloc {
		res |= SbDirPrealloc
	}
	if f.ImagicInodes {
		res |= SbImagicInodes
	}
	if f.HasJournal {
		res |= SbHasJournal
	}
	if f.ExtAttr {
		res |= SbExtAttr
	}
	if f.ResizeInode {
		res |= SbResizeInode
	}
	if f.DirIndex {
		res |= SbDirIndex
	}
	if f.LazyBg {
		res |= SbLazyBg
	}
	if f.ExcludeInode {
		res |= SbExcludeInode
	}
	if f.ExcludeBitmap {
		res |= SbExcludeBitmap
	}
	if f.SparseV2 {
		res |= SbSparseV2
	}

	return res
}

// CompatFeaturesFromInt converts the integer representation of superblock
// compatible features to CompatFeatures struct.
func CompatFeaturesFromInt(f uint32) CompatFeatures {
	return CompatFeatures{
		DirPrealloc:   (f & SbDirPrealloc) > 0,
		ImagicInodes:  (f & SbImagicInodes) > 0,
		HasJournal:    (f & SbHasJournal) > 0,
		ExtAttr:       (f & SbExtAttr) > 0,
		ResizeInode:   (f & SbResizeInode) > 0,
		DirIndex:      (f & SbDirIndex) > 0,
		LazyBg:        (f & SbLazyBg) > 0,
		ExcludeInode:  (f & SbExcludeInode) > 0,
		ExcludeBitmap: (f & SbExcludeBitmap) > 0,
		SparseV2:      (f & SbSparseV2) > 0,
	}
}

// Superblock incompatible features.
const (
	// Compression. Not used.
	SbCompression = 0x1

	// If this is set, then directory entries record the file type. We should use
	// struct ext4_dir_entry_2 for dirents then.
	SbDirentFileType = 0x2

	// Filesystem needs recovery.
	SbRecovery = 0x4

	// Filesystem has a separate journal device.
	SbJournalDev = 0x8

	// Meta block groups. Moves the group descriptors from the congested first
	// block group into the first group of each metablock group to increase the
	// maximum block groups limit and hence support much larger filesystems.
	//
	// See https://www.kernel.org/doc/html/latest/filesystems/ext4/overview.html#meta-block-groups.
	SbMetaBG = 0x10

	// This filesystem uses extents. Must be set in ext4 filesystems.
	SbExtents = 0x40

	// Marks that this filesystem addresses blocks with 64-bits. Hence can support
	// 2^64 data blocks.
	SbIs64Bit = 0x80

	// Multiple mount protection.
	SbMMP = 0x100

	// Flexible block groups. Several block groups are tied into one logical block
	// group so that all the metadata for the block groups (bitmaps and inode
	// tables) are close together for faster loading. Consequently, large files
	// will be continuous on disk. However, this does not affect the placement of
	// redundant superblocks and group descriptors.
	//
	// See https://www.kernel.org/doc/html/latest/filesystems/ext4/overview.html#flexible-block-groups.
	SbFlexBg = 0x200

	// Inodes can store extended attributes when they get really large.
	SbExtAttrInode = 0x400

	// Data in directory entry. Not used.
	SbDirData = 0x1000

	// Metadata checksum seed is stored in the superblock. Enables the admin to
	// change the UUID of a metadata_csum filesystem when mounted.
	SbCsumSeed = 0x2000

	// Large directory enabled. Directory htree can be 3 levels deep.
	// Directory htrees are allowed to be 2 levels deep otherwise.
	SbLargeDir = 0x4000

	// Allows inline data in inodes for really small files.
	SbInlineData = 0x8000

	// This fs contains encrypted inodes.
	SbEncrypted = 0x10000
)

// IncompatFeatures represents a superblock's incompatible feature set. If the
// kernel does not understand any of these feature, it should refuse to mount.
type IncompatFeatures struct {
	Compression    bool
	DirentFileType bool
	Recovery       bool
	JournalDev     bool
	MetaBG         bool
	Extents        bool
	Is64Bit        bool
	MMP            bool
	FlexBg         bool
	ExtAttrInode   bool
	DirData        bool
	CsumSeed       bool
	LargeDir       bool
	InlineData     bool
	Encrypted      bool
}

// ToInt converts superblock incompatible features back to its 32-bit rep.
func (f IncompatFeatures) ToInt() uint32 {
	var res uint32

	if f.Compression {
		res |= SbCompression
	}
	if f.DirentFileType {
		res |= SbDirentFileType
	}
	if f.Recovery {
		res |= SbRecovery
	}
	if f.JournalDev {
		res |= SbJournalDev
	}
	if f.MetaBG {
		res |= SbMetaBG
	}
	if f.Extents {
		res |= SbExtents
	}
	if f.Is64Bit {
		res |= SbIs64Bit
	}
	if f.MMP {
		res |= SbMMP
	}
	if f.FlexBg {
		res |= SbFlexBg
	}
	if f.ExtAttrInode {
		res |= SbExtAttrInode
	}
	if f.DirData {
		res |= SbDirData
	}
	if f.CsumSeed {
		res |= SbCsumSeed
	}
	if f.LargeDir {
		res |= SbLargeDir
	}
	if f.InlineData {
		res |= SbInlineData
	}
	if f.Encrypted {
		res |= SbEncrypted
	}

	return res
}

// IncompatFeaturesFromInt converts the integer representation of superblock
// incompatible features to IncompatFeatures struct.
func IncompatFeaturesFromInt(f uint32) IncompatFeatures {
	return IncompatFeatures{
		Compression:    (f & SbCompression) > 0,
		DirentFileType: (f & SbDirentFileType) > 0,
		Recovery:       (f & SbRecovery) > 0,
		JournalDev:     (f & SbJournalDev) > 0,
		MetaBG:         (f & SbMetaBG) > 0,
		Extents:        (f & SbExtents) > 0,
		Is64Bit:        (f & SbIs64Bit) > 0,
		MMP:            (f & SbMMP) > 0,
		FlexBg:         (f & SbFlexBg) > 0,
		ExtAttrInode:   (f & SbExtAttrInode) > 0,
		DirData:        (f & SbDirData) > 0,
		CsumSeed:       (f & SbCsumSeed) > 0,
		LargeDir:       (f & SbLargeDir) > 0,
		InlineData:     (f & SbInlineData) > 0,
		Encrypted:      (f & SbEncrypted) > 0,
	}
}

// Superblock readonly compatible features.
const (
	// Sparse superblocks. Only groups with number either 0 or a power of 3, 5, or
	// 7 will have redundant copies of the superblock and block descriptors.
	SbSparse = 0x1

	// This fs has been used to store a file >= 2GiB.
	SbLargeFile = 0x2

	// Btree directory structure. Not used.
	SbBTreeDir = 0x4

	// This fs contains files whose sizes are represented in units of logicals
	// blocks, not 512-byte sectors.
	SbHugeFile = 0x8

	// Group descriptors have checksums.
	SbGdtCsum = 0x10

	// Ext3 has a 32,000 subdirectory limit. This tells that the limit no longer
	// applies. New limit is 64,999.
	SbDirNlink = 0x20

	// Indicates that large inodes exist on this filesystem.
	SbExtraIsize = 0x40

	// Indicates the existence of a snapshot.
	SbHasSnapshot = 0x80

	// Enables usage tracking for all quota types.
	SbQuota = 0x100

	// When the `bigalloc` feature is set, the minimum allocation unit becomes a
	// cluster rather than a data block. Then block bitmaps track clusters, not
	// data blocks.
	//
	// See https://www.kernel.org/doc/html/latest/filesystems/ext4/overview.html#bigalloc.
	SbBigalloc = 0x200

	// Fs supports metadata checksumming.
	SbMetadataCsum = 0x400

	// Fs supports replicas. Not used.
	SbReplica = 0x800

	// Marks this filesystem as readonly. Should refuse to mount in read/write mode.
	SbReadOnly = 0x1000

	// Fs tracks project quotas. Not used.
	SbProject = 0x2000
)

// RoCompatFeatures represents a superblock's readonly compatible feature set.
// If the kernel does not understand any of these feature, it can still mount
// readonly. But if the user wants to mount read/write, the kernel should
// refuse to mount.
type RoCompatFeatures struct {
	Sparse       bool
	LargeFile    bool
	BTreeDir     bool
	HugeFile     bool
	GdtCsum      bool
	DirNlink     bool
	ExtraIsize   bool
	HasSnapshot  bool
	Quota        bool
	Bigalloc     bool
	MetadataCsum bool
	Replica      bool
	ReadOnly     bool
	Project      bool
}

// ToInt converts superblock readonly compatible features to its 32-bit rep.
func (f RoCompatFeatures) ToInt() uint32 {
	var res uint32

	if f.Sparse {
		res |= SbSparse
	}
	if f.LargeFile {
		res |= SbLargeFile
	}
	if f.BTreeDir {
		res |= SbBTreeDir
	}
	if f.HugeFile {
		res |= SbHugeFile
	}
	if f.GdtCsum {
		res |= SbGdtCsum
	}
	if f.DirNlink {
		res |= SbDirNlink
	}
	if f.ExtraIsize {
		res |= SbExtraIsize
	}
	if f.HasSnapshot {
		res |= SbHasSnapshot
	}
	if f.Quota {
		res |= SbQuota
	}
	if f.Bigalloc {
		res |= SbBigalloc
	}
	if f.MetadataCsum {
		res |= SbMetadataCsum
	}
	if f.Replica {
		res |= SbReplica
	}
	if f.ReadOnly {
		res |= SbReadOnly
	}
	if f.Project {
		res |= SbProject
	}

	return res
}

// RoCompatFeaturesFromInt converts the integer representation of superblock
// readonly compatible features to RoCompatFeatures struct.
func RoCompatFeaturesFromInt(f uint32) RoCompatFeatures {
	return RoCompatFeatures{
		Sparse:       (f & SbSparse) > 0,
		LargeFile:    (f & SbLargeFile) > 0,
		BTreeDir:     (f & SbBTreeDir) > 0,
		HugeFile:     (f & SbHugeFile) > 0,
		GdtCsum:      (f & SbGdtCsum) > 0,
		DirNlink:     (f & SbDirNlink) > 0,
		ExtraIsize:   (f & SbExtraIsize) > 0,
		HasSnapshot:  (f & SbHasSnapshot) > 0,
		Quota:        (f & SbQuota) > 0,
		Bigalloc:     (f & SbBigalloc) > 0,
		MetadataCsum: (f & SbMetadataCsum) > 0,
		Replica:      (f & SbReplica) > 0,
		ReadOnly:     (f & SbReadOnly) > 0,
		Project:      (f & SbProject) > 0,
	}
}

// OSCode represents ext4 operating system codes.
type OSCode uint32

// Different operating system codes.
const (
	Linux   OSCode = 0
	Hurd    OSCode = 1
	Masix   OSCode = 2
	FreeBSD OSCode = 3
	Lites   OSCode = 4
)

// SbErrorPolicy represents the superblock error policy.
type SbErrorPolicy uint16

// The superblock error policy is one of the following.
const (
	Continue        SbErrorPolicy = 1
	RemountReadOnly SbErrorPolicy = 2
	Panic           SbErrorPolicy = 3
)

// These are the different super block states.
const (
	// Cleanly umounted.
	SbUmounted = 0x1

	// Errors detected
	SbError = 0x2

	// Orphans being recovered.
	SbOrphanRecovery = 0x4
)

// SbState represents all the different combinations of the super block state.
type SbState struct {
	Umounted       bool
	Error          bool
	OrphanRecovery bool
}

// ToInt converts a SbState struct back to its 16-bit representation.
func (s SbState) ToInt() uint16 {
	var res uint16

	if s.Umounted {
		res |= SbUmounted
	}
	if s.Error {
		res |= SbError
	}
	if s.OrphanRecovery {
		res |= SbOrphanRecovery
	}

	return res
}

// SbStateFromInt converts the 16-bit state representation to a SbState struct.
func SbStateFromInt(state uint16) SbState {
	return SbState{
		Umounted:       (state & SbUmounted) > 0,
		Error:          (state & SbError) > 0,
		OrphanRecovery: (state & SbOrphanRecovery) > 0,
	}
}
