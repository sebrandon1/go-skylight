package lib

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RedemptionEvent is emitted by RewardsPoller whenever a reward is newly
// redeemed. It marshals cleanly to JSON for downstream consumers.
type RedemptionEvent struct {
	// RewardID is the Skylight reward identifier.
	RewardID string `json:"reward_id"`
	// RewardName is the human-readable reward title.
	RewardName string `json:"reward_name"`
	// ChildName is the name of the family member who redeemed the reward, or
	// empty if the category cannot be resolved.
	ChildName string `json:"child_name,omitempty"`
	// CategoryID is the raw Skylight category identifier.
	CategoryID string `json:"category_id,omitempty"`
	// Points is the point value of the redeemed reward.
	Points int `json:"points"`
	// ObservedAt is the wall-clock time when the poller first observed the
	// redemption. It is not the server-side redemption timestamp.
	ObservedAt time.Time `json:"observed_at"`
}

// pollerState is the on-disk deduplication state.
type pollerState struct {
	SeenRewardIDs []string `json:"seen_reward_ids"`
}

// RewardsPoller polls a Skylight frame for newly redeemed rewards and emits
// RedemptionEvent values on a channel. Start it with Start and stop it with
// Stop. Events() returns the read-only event channel.
//
// Deduplication is persisted to a local JSON file (default
// ~/.skylight/poller-state.json) so restarts do not double-fire events.
type RewardsPoller struct {
	client    *Client
	frameID   string
	interval  time.Duration
	stateFile string
	logger    *slog.Logger

	events chan RedemptionEvent
	stop   chan struct{}
	done   chan struct{}

	mu   sync.Mutex
	seen map[string]struct{}
}

// NewRewardsPoller constructs a RewardsPoller. stateFile may be empty, in
// which case it defaults to ~/.skylight/poller-state.json.
func NewRewardsPoller(client *Client, frameID string, interval time.Duration, stateFile string) *RewardsPoller {
	var homeDirErr error
	if stateFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			homeDirErr = err
			home = os.TempDir()
		}
		stateFile = filepath.Join(home, ".skylight", "poller-state.json")
	}
	p := &RewardsPoller{
		client:    client,
		frameID:   frameID,
		interval:  interval,
		stateFile: stateFile,
		logger:    client.logger,
		events:    make(chan RedemptionEvent, 64),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		seen:      make(map[string]struct{}),
	}
	if homeDirErr != nil && p.logger != nil {
		p.logger.Warn("poller: home directory unavailable, state file in temp dir and may be lost on reboot",
			slog.String("path", stateFile), slog.String("error", homeDirErr.Error()))
	}
	p.loadState()
	return p
}

// Events returns the read-only channel on which RedemptionEvent values are
// delivered. The channel is buffered (capacity 64). Consumers should drain it
// promptly to avoid blocking the poll loop.
func (p *RewardsPoller) Events() <-chan RedemptionEvent {
	return p.events
}

// Start begins polling in a new goroutine. ctx is used for long-lived
// cancellation (e.g. signal.NotifyContext). The poller also respects Stop().
func (p *RewardsPoller) Start(ctx context.Context) {
	go p.loop(ctx)
}

// Stop signals the poll loop to exit and waits for it to finish.
func (p *RewardsPoller) Stop() {
	close(p.stop)
	<-p.done
}

func (p *RewardsPoller) loop(ctx context.Context) {
	defer close(p.done)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Poll once immediately on start.
	p.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stop:
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *RewardsPoller) poll(ctx context.Context) {
	var (
		childNames map[string]string
		rewards    []Reward
		rewardErr  error
		wg         sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		childNames = p.resolveChildNames()
	}()
	go func() {
		defer wg.Done()
		rewards, rewardErr = p.client.ListRewards(p.frameID)
	}()
	wg.Wait()

	if rewardErr != nil {
		if p.logger != nil {
			p.logger.Warn("poller: ListRewards failed", slog.String("error", rewardErr.Error()))
		}
		return
	}

	// Bound seen-state to currently redeemed IDs only. IDs for rewards that
	// are no longer redeemed are dropped so memory and the on-disk file stay
	// O(active redemptions) instead of growing forever.
	current := make(map[string]struct{})
	var newlyRedeemed []Reward
	p.mu.Lock()
	for _, r := range rewards {
		if !r.Redeemed {
			continue
		}
		current[r.ID] = struct{}{}
		if _, already := p.seen[r.ID]; !already {
			newlyRedeemed = append(newlyRedeemed, r)
		}
	}
	changed := !sameStringSet(p.seen, current)
	p.seen = current
	p.mu.Unlock()

	for _, r := range newlyRedeemed {
		event := RedemptionEvent{
			RewardID:   r.ID,
			RewardName: r.Title,
			CategoryID: r.CategoryID,
			ChildName:  childNames[r.CategoryID],
			Points:     r.Points,
			ObservedAt: time.Now(),
		}

		select {
		case <-ctx.Done():
			return
		case p.events <- event:
		default:
			// Channel full; log and drop rather than block the poll loop.
			if p.logger != nil {
				p.logger.Warn("poller: event channel full, dropping event",
					slog.String("reward_id", r.ID),
				)
			}
		}
	}

	if changed || len(newlyRedeemed) > 0 {
		p.saveState()
	}
}

// sameStringSet reports whether a and b contain the same keys.
func sameStringSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func (p *RewardsPoller) resolveChildNames() map[string]string {
	categories, err := p.client.ListCategories(p.frameID)
	if err != nil {
		return nil
	}
	m := make(map[string]string, len(categories))
	for _, c := range categories {
		m[c.ID] = c.Name
	}
	return m
}

func (p *RewardsPoller) loadState() {
	data, err := os.ReadFile(p.stateFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) && p.logger != nil {
			p.logger.Warn("poller: could not read state file", slog.String("path", p.stateFile), slog.String("error", err.Error()))
		}
		return
	}
	var state pollerState
	if err := json.Unmarshal(data, &state); err != nil {
		if p.logger != nil {
			p.logger.Warn("poller: corrupt state file, starting fresh", slog.String("error", err.Error()))
		}
		return
	}
	for _, id := range state.SeenRewardIDs {
		p.seen[id] = struct{}{}
	}
}

func (p *RewardsPoller) saveState() {
	p.mu.Lock()
	ids := make([]string, 0, len(p.seen))
	for id := range p.seen {
		ids = append(ids, id)
	}
	p.mu.Unlock()

	state := pollerState{SeenRewardIDs: ids}
	data, err := json.Marshal(state)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("poller: failed to marshal state", slog.String("error", err.Error()))
		}
		return
	}

	if err := os.MkdirAll(filepath.Dir(p.stateFile), 0o700); err != nil {
		if p.logger != nil {
			p.logger.Warn("poller: failed to create state directory", slog.String("path", filepath.Dir(p.stateFile)), slog.String("error", err.Error()))
		}
		return
	}
	if err := os.WriteFile(p.stateFile, data, 0o600); err != nil {
		if p.logger != nil {
			p.logger.Warn("poller: failed to write state file", slog.String("path", p.stateFile), slog.String("error", err.Error()))
		}
	}
}
