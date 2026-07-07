package tasks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/cmd/filediver-gui/tasks"
)

func TestTasks(t *testing.T) {
	require := require.New(t)

	task := tasks.Pipeline(
		"in->a.x", true, "->a.someflag",
		"##a##", func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error) {
			x := params["x"].(int)
			someflag := params["someflag"] == true
			return map[string]any{"x": x + 1, "f": someflag}, nil
		},
		"a.x->b.x", "a.f->f",
		"##b##", func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error) {
			x := params["x"].(int)
			return map[string]any{"x": x * 2}, nil
		},
		"b.x->out",
	)
	{
		st := task.Run(context.Background(), map[string]any{"in": 2})
		require.True(st.Done)
		require.True(st.JustFinished)
		require.NoError(st.Err)
		require.Equal(1.0, st.Prog)
		require.Equal(map[string]any{"out": 6, "f": true}, st.Res)
	}
	{
		ch := make(chan tasks.TaskState)
		task.GoCh(ch, context.Background(), map[string]any{"in": 2})
		var st tasks.TaskState
		for st = range ch {
		}
		require.True(st.Done)
		require.True(st.JustFinished)
		require.NoError(st.Err)
		require.Equal(1.0, st.Prog)
		require.Equal(map[string]any{"out": 6, "f": true}, st.Res)
	}
}
