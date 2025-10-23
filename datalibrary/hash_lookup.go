package datalib

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type hashLookup0x7056 struct {
	ParentCount            uint32
	Parents                []hashLookupParent
	HashCount1             uint32
	Hashes1                []stingray.Hash
	HashMap1EntryCount     uint32
	HashMap1               []hashLookupMapEntry
	HashCount2             uint32
	Hashes2                []stingray.Hash
	UnknownTypeIndicator   uint32
	Hashes2MappingCount    uint32
	Hashes2Mapping         []hashLookupHashMapping
	ThinHashMap1EntryCount uint32
	ThinHashMap1           []hashLookupThinMapEntry
	HashCount3             uint32
	Hashes3                []stingray.Hash
	HashMap2EntryCount     uint32
	HashMap2               []hashLookupMapEntry
	LookupTreeCount1       uint32
	LookupTrees1           []hashLookupTree
	HashMap3EntryCount     uint32
	HashMap3               []hashLookupMapEntry
	LookupTreeCount2       uint32
	LookupTrees2           []hashLookupTree
	HashMap4EntryCount     uint32
	HashMap4               []hashLookupMapEntry
	LookupTreeCount3       uint32
	LookupTrees3           []hashLookupTree
	HashMap5EntryCount     uint32
	HashMap5               []hashLookupMapEntry
	LookupTreeCount4       uint32
	LookupTrees4           []hashLookupTree
	DEADBEE7               uint32
}

type hashLookupParent struct {
	ItemCount uint32
	Items     []stingray.Hash
}

type hashLookupMapEntry struct {
	Key   uint64
	Value uint64
}

type hashLookupThinMapEntry struct {
	Hash  uint32
	Index uint32
}

type hashLookupHashMapping struct {
	Type  uint32
	Index uint32
	Count uint32
}

type hashLookupTree struct {
	Type       stingray.ThinHash
	UnkInt     uint32
	EntryCount uint32
	Entries    []hashLookupMapEntry
}

func parseHashLookup(r io.Reader) (map[uint64]stingray.Hash, error) {
	var hashLookup hashLookup0x7056
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ParentCount); err != nil {
		return nil, err
	}

	hashLookup.Parents = make([]hashLookupParent, 0)
	for i := uint32(0); i < hashLookup.ParentCount; i++ {
		var count uint32 = 0
		for count == 0 {
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, err
			}
		}
		items := make([]stingray.Hash, count)
		if err := binary.Read(r, binary.LittleEndian, &items); err != nil {
			return nil, err
		}
		hashLookup.Parents = append(hashLookup.Parents, hashLookupParent{
			ItemCount: count,
			Items:     items,
		})
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount1); err != nil {
		return nil, fmt.Errorf("read hash_count_01 %v", err)
	}
	hashLookup.Hashes1 = make([]stingray.Hash, hashLookup.HashCount1)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes1); err != nil {
		return nil, fmt.Errorf("read hashes_01 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap1EntryCount); err != nil {
		return nil, fmt.Errorf("read hash_map_01_entry_count %v", err)
	}
	hashLookup.HashMap1 = make([]hashLookupMapEntry, hashLookup.HashMap1EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap1); err != nil {
		return nil, fmt.Errorf("read hash_map_01 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount2); err != nil {
		return nil, fmt.Errorf("read hash_count_02 %v", err)
	}
	hashLookup.Hashes2 = make([]stingray.Hash, hashLookup.HashCount2)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2); err != nil {
		return nil, fmt.Errorf("read hashes_02 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.UnknownTypeIndicator); err != nil {
		return nil, fmt.Errorf("read unk_type_ind %v", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2MappingCount); err != nil {
		return nil, fmt.Errorf("read hashes_02_mapping_count %v", err)
	}
	hashLookup.Hashes2Mapping = make([]hashLookupHashMapping, hashLookup.Hashes2MappingCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2Mapping); err != nil {
		return nil, fmt.Errorf("read hashes_02_mapping %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ThinHashMap1EntryCount); err != nil {
		return nil, fmt.Errorf("read thin_hash_map_01_entry_count %v", err)
	}
	hashLookup.ThinHashMap1 = make([]hashLookupThinMapEntry, hashLookup.ThinHashMap1EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ThinHashMap1); err != nil {
		return nil, fmt.Errorf("read thin_hash_map_01 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount3); err != nil {
		return nil, fmt.Errorf("read hash_count_03 %v", err)
	}
	hashLookup.Hashes3 = make([]stingray.Hash, hashLookup.HashCount3)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes3); err != nil {
		return nil, fmt.Errorf("read hashes_03 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap2EntryCount); err != nil {
		return nil, fmt.Errorf("read hash_map_02_entry_count %v", err)
	}
	hashLookup.HashMap2 = make([]hashLookupMapEntry, hashLookup.HashMap2EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap2); err != nil {
		return nil, fmt.Errorf("read hash_map_02 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount1); err != nil {
		return nil, fmt.Errorf("read lookup_tree_count_01 %v", err)
	}
	hashLookup.LookupTrees1 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount1; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, fmt.Errorf("read lookup_trees_01 type %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, fmt.Errorf("read lookup_trees_01 unkint %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, fmt.Errorf("read lookup_trees_01 entry_count %v", err)
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, fmt.Errorf("read lookup_trees_01 entries %v", err)
		}
		hashLookup.LookupTrees1 = append(hashLookup.LookupTrees1, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap3EntryCount); err != nil {
		return nil, fmt.Errorf("read hash_map_03_entry_count %v", err)
	}
	hashLookup.HashMap3 = make([]hashLookupMapEntry, hashLookup.HashMap3EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap3); err != nil {
		return nil, fmt.Errorf("read hash_map_03 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount2); err != nil {
		return nil, fmt.Errorf("read lookup_tree_count_02 %v", err)
	}
	hashLookup.LookupTrees2 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount2; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, fmt.Errorf("read lookup_trees_02 type %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, fmt.Errorf("read lookup_trees_02 unkint %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, fmt.Errorf("read lookup_trees_02 entry_count %v", err)
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, fmt.Errorf("read lookup_trees_02 entries %v", err)
		}
		hashLookup.LookupTrees2 = append(hashLookup.LookupTrees2, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap4EntryCount); err != nil {
		return nil, fmt.Errorf("read hash_map_04_entry_count %v", err)
	}
	hashLookup.HashMap4 = make([]hashLookupMapEntry, hashLookup.HashMap4EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap4); err != nil {
		return nil, fmt.Errorf("read hash_map_04 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount3); err != nil {
		return nil, fmt.Errorf("read lookup_tree_count_03 %v", err)
	}
	hashLookup.LookupTrees3 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount3; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, fmt.Errorf("read lookup_trees_03 type %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, fmt.Errorf("read lookup_trees_03 unkint %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, fmt.Errorf("read lookup_trees_03 entry_count %v", err)
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, fmt.Errorf("read lookup_trees_03 entries %v", err)
		}
		hashLookup.LookupTrees3 = append(hashLookup.LookupTrees3, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap5EntryCount); err != nil {
		return nil, fmt.Errorf("read hash_map_05_entry_count %v", err)
	}
	hashLookup.HashMap5 = make([]hashLookupMapEntry, hashLookup.HashMap5EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap5); err != nil {
		return nil, fmt.Errorf("read hash_map_05 %v", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount4); err != nil {
		return nil, fmt.Errorf("read lookup_tree_count_04 %v", err)
	}
	hashLookup.LookupTrees4 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount4; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, fmt.Errorf("read lookup_trees_04 type %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, fmt.Errorf("read lookup_trees_04 unkint %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, fmt.Errorf("read lookup_trees_04 entry_count %v", err)
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, fmt.Errorf("read lookup_trees_04 entries %v", err)
		}
		hashLookup.LookupTrees4 = append(hashLookup.LookupTrees4, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.DEADBEE7); err != nil {
		return nil, fmt.Errorf("read deadbeet %v", err)
	}

	if hashLookup.DEADBEE7 != 0xDEADBEE7 {
		return nil, fmt.Errorf("invalid format for 0x7056bc19c69f0f07.hash_lookup, expected final bytes read to be 0xDEADBEE7 but were %#08x", hashLookup.DEADBEE7)
	}

	toReturn := make(map[uint64]stingray.Hash)
	for _, entry := range hashLookup.HashMap5 {
		if entry.Key == 0x0 {
			continue
		}
		toReturn[entry.Key] = stingray.Hash{Value: entry.Value}
	}
	return toReturn, nil
}
