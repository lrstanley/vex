// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package tasks

import (
	"context"
	"maps"
	"sync"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

var _ types.TaskState = &state{} // Ensure state implements types.TaskState.

type state struct {
	mu    sync.RWMutex
	tasks map[string]*types.Task
}

func NewState() types.TaskState {
	return &state{
		tasks: make(map[string]*types.Task),
	}
}

func (s *state) AddTask(id string, metadata map[string]any, cancel context.CancelFunc) tea.Cmd {
	return func() tea.Msg {
		s.mu.Lock()
		defer s.mu.Unlock()

		if metadata == nil {
			metadata = make(map[string]any)
		}

		s.tasks[id] = &types.Task{
			ID:       id,
			Metadata: metadata,
			Cancel:   cancel,
		}

		return types.TaskMsg{Msg: types.TaskStartedMsg{
			ID:       id,
			Metadata: metadata,
		}}
	}
}

func (s *state) Run(id string, metadata map[string]any, fn func(context.Context) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		nctx, cancel := context.WithCancel(context.Background())
		f := func() tea.Msg {
			defer cancel()
			return fn(nctx)
		}
		return tea.Batch(
			s.AddTask(id, metadata, cancel),
			f,
		)
	}
}

func (s *state) CancelTask(id string) tea.Cmd {
	return func() tea.Msg {
		s.mu.Lock()
		defer s.mu.Unlock()

		task, exists := s.tasks[id]
		if !exists {
			return nil
		}

		if task.Cancel != nil {
			task.Cancel()
		}

		delete(s.tasks, id)

		return types.TaskMsg{Msg: types.TaskCancelledMsg{
			ID:       id,
			Metadata: task.Metadata,
		}}
	}
}

func (s *state) CancelTasksByFilter(filters map[string]any) tea.Cmd {
	s.mu.Lock()
	defer s.mu.Unlock()

	var tasksToCancel []*types.Task

	for _, task := range s.tasks {
		if s.taskMatchesFilters(task, filters) {
			tasksToCancel = append(tasksToCancel, task)
		}
	}

	var cmds []tea.Cmd
	for _, task := range tasksToCancel {
		cmds = append(cmds, func() tea.Msg {
			s.mu.Lock()
			defer s.mu.Unlock()

			if task.Cancel != nil {
				task.Cancel()
			}

			delete(s.tasks, task.ID)

			return types.TaskMsg{Msg: types.TaskCancelledMsg{
				ID:       task.ID,
				Metadata: task.Metadata,
			}}
		})
	}

	return tea.Batch(cmds...)
}

func (s *state) taskMatchesFilters(task *types.Task, filters map[string]any) bool {
	for key, expectedValue := range filters {
		if actualValue, exists := task.Metadata[key]; !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}

func (s *state) GetTask(id string) *types.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *state) GetAllTasks() map[string]*types.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*types.Task)
	maps.Copy(result, s.tasks)
	return result
}

func (s *state) GetTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

func (s *state) CancelAll() tea.Cmd {
	var cmds []tea.Cmd

	for _, task := range s.tasks {
		cmds = append(cmds, func() tea.Msg {
			s.mu.Lock()
			defer s.mu.Unlock()

			if task.Cancel != nil {
				task.Cancel()
			}

			delete(s.tasks, task.ID)

			return types.TaskMsg{Msg: types.TaskCancelledMsg{
				ID:       task.ID,
				Metadata: task.Metadata,
			}}
		})
	}

	return tea.Batch(cmds...)
}
