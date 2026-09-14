/**
 * Lightweight, Medium Cinematic Scroll Controller
 * - Silky-smooth top scroll progress indicator
 * - Apple Dynamic Island nav elevation on scroll
 * - Non-interfering: leaves card 3D transforms to CSS & tilt engine
 */

if (typeof window !== 'undefined') {
  const progressBar = document.querySelector<HTMLElement>('.scroll-progress');
  const nav = document.querySelector<HTMLElement>('.dynamic-island');

  let ticking = false;

  function updateScrollCinematics() {
    const scrollY = window.scrollY;
    const maxScroll = document.documentElement.scrollHeight - window.innerHeight;

    // 1. Sleek Top Scroll Progress Bar
    if (progressBar && maxScroll > 0) {
      const progress = Math.min(1, Math.max(0, scrollY / maxScroll));
      progressBar.style.transform = `scaleX(${progress})`;
      progressBar.style.opacity = scrollY > 30 ? '1' : '0';
    }

    // 2. Dynamic Island Nav elevation (medium, subtle scale)
    if (nav) {
      if (scrollY > 40) {
        nav.style.transform = 'scale(0.99)';
        nav.style.boxShadow = '0 18px 38px -10px rgba(0,0,0,0.92), 0 0 20px rgba(255,255,255,0.04)';
        nav.style.borderColor = 'rgba(255, 255, 255, 0.18)';
      } else {
        nav.style.transform = 'scale(1)';
        nav.style.boxShadow = '0 16px 36px -10px rgba(0,0,0,0.9), 0 0 20px rgba(255,255,255,0.03)';
        nav.style.borderColor = 'rgba(255, 255, 255, 0.12)';
      }
    }

    ticking = false;
  }

  window.addEventListener(
    'scroll',
    () => {
      if (!ticking) {
        requestAnimationFrame(updateScrollCinematics);
        ticking = true;
      }
    },
    { passive: true }
  );

  window.addEventListener('resize', updateScrollCinematics, { passive: true });
  document.addEventListener('DOMContentLoaded', updateScrollCinematics);
  updateScrollCinematics();
}

