// DSAA is the format obtained after extracting the file
// contained in bundles.nxa (which is a DSAR bundle).

#pragma pattern_limit 10000000

// A collection of one or more DSAR chunks.
// Use the orig_archive_file_offsets delta
// between current and next entry to figure
// out how many consecutive chunks should be read.
struct Entry {
    // offset into the to-be-reconstructed archive file
    u64 orig_archive_file_offset;
    // to get the matching chunk in the NXA bundle,
    // find the chunk uncompressed_offset of the same
    // value
    u32 uncompressed_nxa_bundle_offset;
    padding[3];
    // index of bundles.<index>.nxa file
    u8 nxa_bundle_index;
};


// Archive file contained in some other nxa file
struct ArchiveItem {
    u64 size; // size of the archive file
    u32 filename_offset;
    u32 entries_count;
    u32 entries_offset;
    padding[4];
    u8 filename[] @ filename_offset;
    Entry entries[entries_count] @ entries_offset;
};

struct NXAItem {
    u32 filename_offset;
    u8 filename[] @ filename_offset;
};

struct DSAA {
    u8 magic[4];
    u32 unk_00;
    u32 unk_01;
    u32 nxa_item_count;
    u32 archive_item_count;
    padding[4];
    // items are sorted alphabetically by filename
    ArchiveItem archive_items[archive_item_count];
    NXAItem nxa_items[nxa_item_count];
};

DSAA f @ 0;