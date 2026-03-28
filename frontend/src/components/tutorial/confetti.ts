export async function fireCelebrationConfetti(
  reducedMotion: boolean,
): Promise<void> {
  if (reducedMotion) return;
  const confetti = (await import('canvas-confetti')).default;
  const colors = ['#41b1e1', '#7acbee', '#2a8fbd', '#ffffff', '#becad6'];
  const z = 1000000001; // driver.js overlay より上

  // 左右から同時発射（クラッカー風）
  confetti({
    particleCount: 100, angle: 60, spread: 60,
    origin: { x: 0, y: 0.6 }, colors,
    startVelocity: 50, gravity: 1.2, ticks: 250, zIndex: z,
  });
  confetti({
    particleCount: 100, angle: 120, spread: 60,
    origin: { x: 1, y: 0.6 }, colors,
    startVelocity: 50, gravity: 1.2, ticks: 250, zIndex: z,
  });

  // 少し遅れて追加バースト（華やかさアップ）
  setTimeout(() => {
    confetti({
      particleCount: 80, angle: 75, spread: 70,
      origin: { x: 0.1, y: 0.5 }, colors,
      startVelocity: 40, gravity: 1.0, ticks: 200, zIndex: z,
    });
    confetti({
      particleCount: 80, angle: 105, spread: 70,
      origin: { x: 0.9, y: 0.5 }, colors,
      startVelocity: 40, gravity: 1.0, ticks: 200, zIndex: z,
    });
  }, 200);
}
