package extractor

import (
	"fmt"
	"io"
	"slices"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/stingray"
)

type ExtractFunc func(ctx *Context) error

func extractByType(ctx *Context, typ stingray.DataType, extension string) error {
	r, err := ctx.Open(ctx.FileID(), typ)
	if err != nil {
		return err
	}

	var typExtension string
	switch typ {
	case stingray.DataMain:
		typExtension = ".main"
	case stingray.DataStream:
		typExtension = ".stream"
	case stingray.DataGPU:
		typExtension = ".gpu"
	default:
		panic("unhandled case")
	}
	out, err := ctx.CreateFile("." + extension + typExtension)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return nil
}

func extractCombined(ctx *Context, extension string) error {
	if !(ctx.Exists(ctx.FileID(), stingray.DataMain) || ctx.Exists(ctx.FileID(), stingray.DataStream) || ctx.Exists(ctx.FileID(), stingray.DataGPU)) {
		return fmt.Errorf("extractCombined: no data to extract for file")
	}
	out, err := ctx.CreateFile(fmt.Sprintf(".%v", extension))
	if err != nil {
		return err
	}
	defer out.Close()

	for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
		r, err := ctx.Open(ctx.FileID(), typ)
		if err == stingray.ErrFileDataTypeNotExist {
			continue
		}
		if err != nil {
			return err
		}

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
	}

	return nil
}

func ExtractFuncRaw(extension string) ExtractFunc {
	return func(ctx *Context) error {
		for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
			if ctx.Exists(ctx.FileID(), typ) {
				if err := extractByType(ctx, typ, extension); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func ExtractFuncRawSingleType(extension string, typ stingray.DataType) ExtractFunc {
	return func(ctx *Context) error {
		if ctx.Exists(ctx.FileID(), typ) {
			return extractByType(ctx, typ, extension)
		}
		return fmt.Errorf("no %v data found", typ.String())
	}
}

func ExtractFuncRawCombined(extension string) ExtractFunc {
	return func(ctx *Context) error {
		return extractCombined(ctx, extension)
	}
}

// Blender throws a hissy fit if a node is reachable from multiple places in a scene, so we need to remove
// child nodes from the scene before saving.
func ClearChildNodesFromScene(ctx *Context, doc *gltf.Document) {
	nodesToDelete := make([]uint32, 0)
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		ctx.Warnf("No extras in doc? (Should not happen unless nothing was exported)")
		return
	}
	for _, node := range doc.Scenes[0].Nodes {
		nodeMetadata, ok := doc.Nodes[node].Extras.(map[string]any)
		if !ok {
			continue
		}
		hashIface, contains := nodeMetadata["hash"]
		if !contains {
			ctx.Warnf("node %v in scene missing hash information", doc.Nodes[node].Name)
			continue
		}
		hash, ok := hashIface.(string)
		if !ok {
			ctx.Warnf("node %v's hash could not be converted to string", doc.Nodes[node].Name)
			continue
		}
		metadataIface, contains := extras[hash]
		if !contains {
			ctx.Warnf("node %v's metadata was not present in doc extras", doc.Nodes[node].Name)
			continue
		}
		metadata, ok := metadataIface.(map[string]any)
		if !ok {
			ctx.Warnf("node %v's metadata could not be converted", doc.Nodes[node].Name)
			continue
		}
		parentIface, contains := metadata["parent"]
		if !contains {
			ctx.Warnf("node %v in scene missing parent information", doc.Nodes[node].Name)
			continue
		}
		if _, ok := parentIface.(uint32); ok {
			// parent can be converted to uint32, meaning this node is a child node of some other node
			nodesToDelete = append(nodesToDelete, node)
		}
	}
	for _, node := range nodesToDelete {
		idx := slices.Index(doc.Scenes[0].Nodes, node)
		if idx < 0 {
			continue
		}
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes[:idx], doc.Scenes[0].Nodes[idx+1:]...)
	}
}
