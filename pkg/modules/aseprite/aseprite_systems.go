package aseprite

import (
	"time"

	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/yohamta/donburi"
)

type AsepriteSystems struct {
	library *AsepriteLibrary
}

func NewAsepriteSystems(library *AsepriteLibrary) *AsepriteSystems {
	return &AsepriteSystems{
		library: library,
	}
}

func (s *AsepriteSystems) UpdateAnimation(entry *donburi.Entry, dt time.Duration) {
	animator := GetAnimator(entry)
	if animator == nil {
		return
	}

	if animator.State != AnimationStatePlaying {
		return
	}

	library, exists := s.library.animations[images.GetHandle(entry)]
	if !exists {
		return
	}

	tag, exists := library.Meta.Clips[animator.Clip]
	if !exists {
		return
	}

	elapsed := animator.time + dt

	frame := animator.Frame
	frameIndex := tag.From + frame

	duration := time.Duration(library.Frames[frameIndex].Duration) * time.Millisecond
	if duration <= 0 || elapsed < duration {
		animator.time = elapsed
		images.SetFrame(entry, frameIndex)
		return
	}

	nextFrame := frame
	frameCount := tag.To - tag.From + 1

	var completed bool
	for elapsed >= duration {
		elapsed -= duration

		nextFrame, completed = s.getNextFrame(nextFrame, animator, frameCount)
		if completed {
			break
		}

		frameIndex = tag.From + nextFrame
		duration = time.Duration(library.Frames[frameIndex].Duration) * time.Millisecond
	}

	images.SetFrame(entry, frameIndex)

	animator.Frame = nextFrame
	animator.time = elapsed

	if completed {
		animator.State = AnimationStateStopped
	}
}

func (s *AsepriteSystems) getNextFrame(current int, animator *AnimatorModel, frameCount int) (int, bool) {
	current += animator.direction

	loopComplete := current >= frameCount || current < 0
	if !loopComplete {
		return current, false
	}

	if animator.Loops > 0 {
		animator.Loops--
	}

	if animator.Loops == 0 {
		if current >= frameCount {
			current = frameCount - 1
		} else if current < 0 {
			current = 0
		}
		return current, true
	}

	// Loops is -1 (infinite) or still > 0, wrap around
	if animator.direction > 0 {
		current = 0
	} else {
		current = frameCount - 1
	}

	return current, false
}
