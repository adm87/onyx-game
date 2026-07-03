package aseprite

import (
	"time"

	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/yohamta/donburi"
)

type State uint8

type AnimatorModel struct {
	State     State
	Frame     int
	Loops     int
	direction int
	Clip      string
	time      time.Duration
}

type SpriteOptions struct {
	*images.ImageOptions

	State State
	Loops int
	Frame int
	Clip  string
}

type SpriteOption func(*SpriteOptions)

const (
	AnimationStateStopped State = iota
	AnimationStatePlaying
)

var Animator = donburi.NewComponentType[AnimatorModel]()

func DefaultSpriteOptions() *SpriteOptions {
	return &SpriteOptions{
		ImageOptions: images.DefaultImageOptions(),
		State:        AnimationStateStopped,
		Loops:        -1, // -1 for infinite loops
		Frame:        0,
		Clip:         "",
	}
}

func Playing() SpriteOption {
	return func(opts *SpriteOptions) {
		opts.State = AnimationStatePlaying
	}
}

func WithAnimationFrame(frame int) SpriteOption {
	return func(opts *SpriteOptions) {
		opts.Frame = frame
	}
}

func WithLoops(loops int) SpriteOption {
	return func(opts *SpriteOptions) {
		opts.Loops = loops
	}
}

func WithClip(clip string) SpriteOption {
	return func(opts *SpriteOptions) {
		opts.Clip = clip
	}
}

func WithImageOptions(imageOpts ...images.Option) SpriteOption {
	return func(opts *SpriteOptions) {
		for _, opt := range imageOpts {
			opt(opts.ImageOptions)
		}
	}
}

func GetAnimator(entry *donburi.Entry) *AnimatorModel {
	if !entry.HasComponent(Animator) {
		return nil
	}
	return Animator.Get(entry)
}

func SetAnimator(entry *donburi.Entry, info *AnimatorModel) {
	donburi.Add(entry, Animator, info)
}

func GetAnimationState(entry *donburi.Entry) State {
	if !entry.HasComponent(Animator) {
		return AnimationStateStopped
	}
	return Animator.Get(entry).State
}

func SetAnimationState(entry *donburi.Entry, state State) {
	animator := GetAnimator(entry)
	if animator == nil {
		return
	}
	animator.State = state
	SetAnimator(entry, animator)
}

func GetLoops(entry *donburi.Entry) int {
	animator := GetAnimator(entry)
	if animator == nil {
		return 0
	}
	return animator.Loops
}

func SetLoops(entry *donburi.Entry, loops int) {
	animator := GetAnimator(entry)
	if animator == nil {
		return
	}
	animator.Loops = loops
	SetAnimator(entry, animator)
}

func GetAnimationFrame(entry *donburi.Entry) int {
	animator := GetAnimator(entry)
	if animator == nil {
		return 0
	}
	return animator.Frame
}

func SetAnimationFrame(entry *donburi.Entry, frame int) {
	animator := GetAnimator(entry)
	if animator == nil {
		return
	}
	animator.Frame = frame
	SetAnimator(entry, animator)
}

func GetClip(entry *donburi.Entry) string {
	animator := GetAnimator(entry)
	if animator == nil {
		return ""
	}
	return animator.Clip
}

func SetClip(entry *donburi.Entry, clip string) {
	if !entry.HasComponent(Animator) {
		return
	}

	animator := GetAnimator(entry)
	if animator.Clip == clip {
		return
	}

	animator.Clip = clip
	animator.Frame = 0
	animator.State = AnimationStatePlaying
}

func IsIdle(entry *donburi.Entry) bool {
	state := GetAnimationState(entry)
	return state == AnimationStateStopped
}

func IsPlaying(entry *donburi.Entry) bool {
	state := GetAnimationState(entry)
	return state == AnimationStatePlaying
}
