type SoundType = 'info' | 'dispatch' | 'urgent' | 'success';

let audioCtx: AudioContext | null = null;

function getAudioContext(): AudioContext {
  if (!audioCtx) {
    audioCtx = new AudioContext();
  }
  if (audioCtx.state === 'suspended') {
    audioCtx.resume();
  }
  return audioCtx;
}

const SOUND_CONFIGS: Record<
  SoundType,
  { frequencies: number[]; durations: number[]; type: OscillatorType; gain: number }
> = {
  info: {
    frequencies: [523, 659],
    durations: [0.12, 0.12],
    type: 'sine',
    gain: 0.15,
  },
  dispatch: {
    frequencies: [659, 784, 1047],
    durations: [0.1, 0.1, 0.15],
    type: 'sine',
    gain: 0.2,
  },
  urgent: {
    frequencies: [880, 1109, 880, 1109],
    durations: [0.08, 0.08, 0.08, 0.12],
    type: 'triangle',
    gain: 0.25,
  },
  success: {
    frequencies: [523, 659, 784],
    durations: [0.1, 0.1, 0.2],
    type: 'sine',
    gain: 0.12,
  },
};

export function playNotificationSound(type: SoundType = 'info'): void {
  try {
    const ctx = getAudioContext();
    const config = SOUND_CONFIGS[type];
    let startTime = ctx.currentTime;

    config.frequencies.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gainNode = ctx.createGain();

      osc.type = config.type;
      osc.frequency.setValueAtTime(freq, startTime);

      const duration = config.durations[i];
      gainNode.gain.setValueAtTime(0, startTime);
      gainNode.gain.linearRampToValueAtTime(config.gain, startTime + 0.01);
      gainNode.gain.setValueAtTime(config.gain, startTime + duration - 0.03);
      gainNode.gain.linearRampToValueAtTime(0, startTime + duration);

      osc.connect(gainNode);
      gainNode.connect(ctx.destination);

      osc.start(startTime);
      osc.stop(startTime + duration);

      startTime += duration + 0.02;
    });
  } catch {
    // Silently fail if audio is not available
  }
}
