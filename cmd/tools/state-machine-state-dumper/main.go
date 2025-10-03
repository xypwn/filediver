package main

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/state_machine"
)

type GraphMLNode struct {
	ResolvedName            string   `xml:"id,attr"`
	ResolvedAnimationHashes []string `xml:"animations,omitempty"`
}

type GraphMLEdge struct {
	ResolvedEvent string `xml:"desc"`
	Source        string `xml:"source,attr"`
	Target        string `xml:"target,attr"`
}

type GraphMLStateGroup struct {
	EdgeDefault string        `xml:"edgedefault,attr"`
	Nodes       []GraphMLNode `xml:"node"`
	Edges       []GraphMLEdge `xml:"edge"`
}

type GraphMLStateMachine struct {
	XMLName           xml.Name            `xml:"graphml"`
	Namespace         string              `xml:"xmlns,attr"`
	XSINamespace      string              `xml:"xmlns:xsi,attr"`
	XSISchemaLocation string              `xml:"xsi:schemaLocation,attr"`
	Groups            []GraphMLStateGroup `xml:"graph,omitempty"`
}

// Dumping the state machine as a bunch of graphml graphs
// Sorta kinda useful maybe?
func dumpStateMachineStates(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}

	var graphml GraphMLStateMachine
	graphml.Namespace = "http://graphml.graphdrawing.org/xmlns"
	graphml.XSINamespace = "http://www.w3.org/2001/XMLSchema-instance"
	graphml.XSISchemaLocation = "http://graphml.graphdrawing.org/xmlns http://graphml.graphdrawing.org/xmlns/1.0/graphml.xsd"
	graphml.Groups = make([]GraphMLStateGroup, 0)
	for grpIdx, group := range stateMachine.Groups {
		var graph GraphMLStateGroup
		graph.EdgeDefault = "directed"
		graph.Nodes = make([]GraphMLNode, 0)
		graph.Edges = make([]GraphMLEdge, 0)
		for _, state := range group.States {
			node := GraphMLNode{
				ResolvedName:            ctx.LookupHash(state.Name),
				ResolvedAnimationHashes: make([]string, 0),
			}
			for _, hash := range state.AnimationHashes {
				node.ResolvedAnimationHashes = append(node.ResolvedAnimationHashes, ctx.LookupHash(hash))
			}
			graph.Nodes = append(graph.Nodes, node)
			for key, transition := range state.StateTransitions {
				graph.Edges = append(graph.Edges, GraphMLEdge{
					ResolvedEvent: ctx.LookupThinHash(key),
					Source:        ctx.LookupHash(state.Name),
					Target:        ctx.LookupHash(group.States[transition.Index].Name),
				})
			}
		}
		graphml.Groups = []GraphMLStateGroup{graph}

		marshalled, err := xml.MarshalIndent(graphml, "", "    ")
		if err != nil {
			return nil
		}

		writer, err := ctx.CreateFile("/group_" + strconv.FormatUint(uint64(grpIdx), 10) + ".graphml")
		if err != nil {
			return nil
		}
		defer writer.Close()
		writer.Write([]byte(xml.Header))
		writer.Write(marshalled)
	}

	return nil
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	argp := argparse.NewParser("state-machine-state-dumper", "", nil)
	outputDirectory := argp.String("", "output", &argparse.Option{
		Positional: true,
	})
	filenameGlob := argp.String("", "filename_glob", &argparse.Option{
		Positional: true,
	})
	if err := argp.Parse(nil); err != nil {
		prt.Fatalf("argparse: %v", err)
	}

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray.ThinHash{}, func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Animation name dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles(*filenameGlob, "", []string{"state_machine"}, nil, "")
	if err != nil {
		prt.Fatalf("%v", err)
	}

	for fileID := range files {
		name, ok := a.Hashes[fileID.Name]
		if !ok {
			name = fileID.Name.String()
		}
		name = filepath.Base(name)
		var cfg appconfig.Config
		config.InitDefault(&cfg)
		extrCtx, _ := extractor.NewContext(
			ctx,
			fileID,
			a.Hashes,
			a.ThinHashes,
			a.ArmorSets,
			a.SkinOverrideGroups,
			a.LanguageMap,
			a.DataDir,
			nil,
			cfg,
			filepath.Join(*outputDirectory, name),
			[]stingray.Hash{},
			prt.Warnf,
		)
		if err := dumpStateMachineStates(extrCtx); err != nil {
			if errors.Is(err, context.Canceled) {
				prt.NoStatus()
				prt.Warnf("State dump canceled, exiting cleanly")
				return
			} else {
				prt.Errorf("%v", err)
			}
		}
	}
}
