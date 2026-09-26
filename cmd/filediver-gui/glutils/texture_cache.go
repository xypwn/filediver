package glutils

import (
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/stingray"
)

type TextureCacheEntry struct {
	id         uint32
	references uint32
	lastUsed   time.Time
}

type TextureCache struct {
	cache              map[stingray.Hash]map[uint32]TextureCacheEntry
	maxUnusedResidency time.Duration
}

func NewTextureCache(maxDuration time.Duration) *TextureCache {
	return &TextureCache{
		maxUnusedResidency: maxDuration,
	}
}

// Generates a new texture for (hash, target) and adds it to the cache if not already present
func (t *TextureCache) Acquire(hash stingray.Hash, target uint32) (textureId uint32, created bool) {
	if t.cache == nil {
		t.cache = make(map[stingray.Hash]map[uint32]TextureCacheEntry)
	}
	var val TextureCacheEntry
	targets, contains := t.cache[hash]
	if contains {
		val, contains = targets[target]
	} else {
		// If this hash doesn't have a target map, create it
		targets = make(map[uint32]TextureCacheEntry)
	}
	if contains {
		textureId = val.id
		val.lastUsed = time.Now()
		val.references += 1
		fmt.Printf("[cache] Acquiring texture %v (%v): new reference count %v\n", hash.String(), GLTarget(target).String(), val.references)
	} else {
		gl.GenTextures(1, &textureId)
		val = TextureCacheEntry{
			id:         textureId,
			references: 1,
			lastUsed:   time.Now(),
		}
		fmt.Printf("[cache] Created texture %v (%v)\n", hash.String(), GLTarget(target).String())
	}
	created = !contains
	targets[target] = val
	t.cache[hash] = targets
	return
}

func (t *TextureCache) Release(hash stingray.Hash, target uint32) (contains bool) {
	contains = false
	if t.cache == nil {
		return
	}
	var targets map[uint32]TextureCacheEntry
	var val TextureCacheEntry
	if targets, contains = t.cache[hash]; contains {
		if val, contains = targets[target]; contains {
			if val.references > 0 {
				val.references -= 1
			}
			val.lastUsed = time.Now()
			fmt.Printf(
				"[cache] Dereferencing texture %v (%v): new reference count %v\n",
				hash.String(),
				GLTarget(target).String(),
				val.references,
			)
			targets[target] = val
			t.cache[hash] = targets
		}
	}
	return
}

func (t *TextureCache) Delete(hash stingray.Hash, target uint32) (contains bool) {
	contains = false
	if t.cache == nil {
		return
	}
	var targets map[uint32]TextureCacheEntry
	var val TextureCacheEntry
	if targets, contains = t.cache[hash]; contains {
		if val, contains = targets[target]; contains {
			fmt.Printf("[cache] Deleting texture %v target %v\n", hash.String(), GLTarget(target).String())
			delete(t.cache[hash], target)
			gl.DeleteTextures(1, &val.id)
		}
	}
	return
}

func (t *TextureCache) DeleteAll() {
	if t.cache == nil {
		return
	}
	for hash := range t.cache {
		for target := range t.cache[hash] {
			t.Delete(hash, target)
		}
	}
}

func (t *TextureCache) Sweep() {
	if t.cache == nil {
		return
	}
	fmt.Printf("[cache] Sweeping...\n")
	sweepTime := time.Now()
	for hash, targets := range t.cache {
		for target, entry := range targets {
			if entry.references > 0 {
				continue
			}
			entryDuration := sweepTime.Sub(entry.lastUsed)
			fmt.Printf("[cache] Unused texture %v (%v) in residency for %.2fs\n", hash.String(), GLTarget(target).String(), entryDuration.Seconds())
			if entryDuration >= t.maxUnusedResidency {
				t.Delete(hash, target)
			}
		}
	}
	sweepDuration := time.Since(sweepTime)
	fmt.Printf("[cache] Sweep completed in %vms\n", sweepDuration.Milliseconds())
}
