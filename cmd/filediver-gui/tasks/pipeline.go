package tasks

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type subtask struct {
	fn     TaskFunc
	id     string
	name   string
	weight float64
}
type subtaskEdge struct {
	fromName    string
	toSubtaskId string
	toName      string
}
type parsedPipelineArgs struct {
	subtasks  []subtask
	edges     map[string][]subtaskEdge
	constants map[string]any
}

func parsePipelineArgs(args []any) (*parsedPipelineArgs, error) {
	p := &parsedPipelineArgs{
		edges:     make(map[string][]subtaskEdge),
		constants: make(map[string]any),
	}

	existingSubtaskIds := make(map[string]struct{})
	var weightSum float64
	parseSubtask := func(fn TaskFunc, prevArg any) error {
		spec, ok := prevArg.(string)
		if !ok {
			return fmt.Errorf("expected TaskFunc to be preceded by string, but got %T", prevArg)
		}

		var name, id string
		var weight float64
		{
			var sp [3]string
			ok := false
			if j := strings.LastIndex(spec, "##"); j != -1 { // split into last 3 (this way "##" is allowed in name)
				if i := strings.LastIndex(spec[:j], "##"); i != -1 {
					sp = [3]string{spec[:i], spec[i+2 : j], spec[j+2:]}
					ok = true
				}
			}
			if !ok {
				return fmt.Errorf("expected TaskFunc to be preceded by string of form \"<name>##<id>##<weight>\", but got %q", spec)
			}
			name = sp[0]
			id = sp[1]
			if id == "" {
				return fmt.Errorf("parsing id: cannot be empty")
			}
			if slices.Contains([]string{"##in##", "##out##", "##const##"}, id) {
				return fmt.Errorf("invalid id %q (reserved internal ID)", id)
			}
			if len(sp) < 3 || sp[2] == "" {
				weight = 1
			} else {
				var err error
				weight, err = strconv.ParseFloat(sp[2], 64)
				if err != nil {
					return fmt.Errorf("parsing task weight: %w", err)
				}
			}
		}

		if _, exists := existingSubtaskIds[id]; exists {
			return fmt.Errorf("subtask id %q exists more than once (each subtask ID must be unique)", id)
		}
		st := subtask{
			fn:     fn,
			id:     id,
			name:   name,
			weight: weight,
		}
		p.subtasks = append(p.subtasks, st)
		existingSubtaskIds[id] = struct{}{}
		weightSum += weight
		return nil
	}

	for i := len(args) - 1; i >= 0; i-- {
		arg := args[i]
		var prevArg any
		if i >= 1 {
			prevArg = args[i-1]
		}
		switch arg := arg.(type) {
		case string:
			if left, right, ok := strings.Cut(arg, "->"); ok {
				left, right := strings.TrimSpace(left), strings.TrimSpace(right)
				parseVar := func(v, defaultScope string) (subtaskId, name string) {
					if scope, name, ok := strings.Cut(v, "."); ok {
						return scope, name
					} else {
						return defaultScope, v
					}
				}
				var lsid, lname string
				if left != "" {
					// Assignment from variable
					lsid, lname = parseVar(left, "##in##")
				} else {
					// Assignment from constant
					lsid = "##const##"
					lname = fmt.Sprint(len(p.constants))
					p.constants[lname] = prevArg
					i--
				}
				rsid, rname := parseVar(right, "##out##")
				p.edges[lsid] = append(p.edges[lsid], subtaskEdge{lname, rsid, rname})
			} else {
				return nil, fmt.Errorf("expected \"<variable> -> <variable>\", but got %q", arg)
			}
		case TaskFunc:
			if err := parseSubtask(arg, prevArg); err != nil {
				return nil, err
			}
			i--
		case func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error):
			if err := parseSubtask(arg, prevArg); err != nil {
				return nil, err
			}
			i--
		default:
			return nil, fmt.Errorf("unexpected/unknown arg type %T", arg)
		}
	}

	slices.Reverse(p.subtasks) // subtasks were parsed in reverse

	for i := range p.subtasks {
		p.subtasks[i].weight /= weightSum
	}
	return p, nil
}

// args is a list of arbitrary arguments, which are interpreted as a declarative language.
//
// An arg of type TaskFunc (or the underlying func type) is added as a subtask, necessarily
// preceded by a "<name>##<id>##<weight>" string, where <name> is an optional string describing
// the task status prefix, <id> is a unique ID to identify the task and <weight> is an optional
// float (default 1), which describes how significant the progress on this task is to the overall
// progress (weight is automatically distributed between subtasks, so a weight sum of more than
// one is valid and encouraged).
//
// A variable is an input or result value to the whole task (e.g. "url", "path"), or an input
// or result value of a subtask (e.g. "dl.url", "dl.numFiles", where "dl" is the task ID).
//
// Other args specify edges from one variable to another: A value arg followed by "-> <var>"
// (where <var> is the variable) assigns a constant value to a variable. "<var> -> <var>"
// connects one variable to another (e.g. "downloader.output -> extractor.input"). A variable
// without a "." is considered an input/output to the resulting pipeline.
//
// Quick reference:
// - "<name>##<id>##<weight>", TaskFunc: add subtask
// - val, "-> <var>": assign constant
// - "<var> -> <var>": connect variables
func Pipeline(args ...any) TaskFunc {
	p, err := parsePipelineArgs(args)
	if err != nil {
		panic(err.Error())
	}

	return func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (_ map[string]any, err error) {
		varVals := make(map[string]map[string]any)
		propagateVarVals := func(fromId string, values map[string]any) { // propagates given values through edges
			eOut := p.edges[fromId]
			if varVals[fromId] == nil {
				varVals[fromId] = make(map[string]any)
			}
			maps.Copy(varVals[fromId], values)
			for _, e := range eOut {
				if varVals[e.toSubtaskId] == nil {
					varVals[e.toSubtaskId] = make(map[string]any)
				}
				varVals[e.toSubtaskId][e.toName] = varVals[fromId][e.fromName]
			}
		}
		propagateVarVals("##in##", params)
		propagateVarVals("##const##", p.constants)

		var totalProg float64
		runSub := func(sub subtask, params map[string]any) (map[string]any, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			onStatus(sub.name)
			res, err := sub.fn(
				ctx,
				params,
				func(prog float64) {
					onProgress(totalProg + prog*sub.weight)
				},
				func(s string) {
					n := sub.name
					if n != "" {
						n += ": "
					}
					n += s
					onStatus(n)
				},
			)
			if err != nil {
				return nil, err
			}
			totalProg += sub.weight
			onProgress(totalProg)
			return res, err
		}
		for _, sub := range p.subtasks {
			res, err := runSub(sub, varVals[sub.id])
			if err != nil {
				return nil, err
			}
			propagateVarVals(sub.id, res)
		}
		return varVals["##out##"], nil
	}
}
