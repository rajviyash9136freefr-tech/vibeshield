/**
 * Apple-Grade Bidirectional Cinematic Scroll Engine
 *  - Handles Upper and Lower cinematic scroll transforms
 *  - Kinetic marquee crawl speedup on scroll momentum
 *  - Dynamic Island nav scale on scroll
 *  - Continuous RAF loop with lerp easing
 */

const reduced = document.documentElement.dataset.reducedMotion === '1';

if (!reduced && typeof window !== 'undefined') {
  const elements = document.querySelectorAll<HTMLElement>('[data-reveal], [data-cinematic]');
  const crawlTracks = document.querySelectorAll<HTMLElement>('.crawl-track');
  const nav = document.querySelector<HTMLElement>('.dynamic-island');

  let lastScrollY = window.scrollY;
  let scrollVelocity = 0;
  let isTicking = false;

  function onScroll() {
    const currentScrollY = window.scrollY;
    scrollVelocity = Math.abs(currentScrollY - lastScrollY);
    lastScrollY = currentScrollY;

    if (!isTicking) {
      requestAnimationFrame(updateCinematics);
      isTicking = true;
    }
  }

  function updateCinematics() {
    const vh = window.innerHeight;
    const scrollY = window.scrollY;

    // 1. Dynamic Island Nav Transform on scroll
    if (nav) {
      if (scrollY > 50) {
        nav.style.transform = 'scale(0.96)';
        nav.style.boxShadow = '0 20px 40px -10px rgba(0,0,0,0.95), 0 0 25px rgba(255,255,255,0.06)';
        nav.style.borderColor = 'rgba(255, 255, 255, 0.2)';
      } else {
        nav.style.transform = 'scale(1)';
        nav.style.boxShadow = '0 16px 36px -10px rgba(0,0,0,0.9), 0 0 20px rgba(255,255,255,0.03)';
        nav.style.borderColor = 'rgba(255, 255, 255, 0.12)';
      }
    }

    // 2. Bidirectional Cinematic Scroll Reveals (Upper & Lower)
    elements.forEach((el) => {
      const rect = el.getBoundingClientRect();
      const top = rect.top;
      const bottom = rect.bottom;

      // When element is in viewport with generous buffer
      if (top < vh * 0.92 && bottom > 0) {
        el.classList.add('revealed');
        // Progressive lift & settle
        const progress = Math.min(1, Math.max(0, (vh - top) / (vh * 0.4)));
        const translateY = (1 - progress) * 24;
        const scale = 0.97 + progress * 0.03;
        el.style.transform = `translate3d(0, ${translateY.toFixed(1)}px, 0) scale(${scale.toFixed(3)})`;
        el.style.opacity = `${Math.min(1, 0.2 + progress * 0.8).toFixed(2)}`;
      } else if (top >= vh * 0.95) {
        // Lower cinematic: scrolled down below viewport, prepare entrance
        el.classList.remove('revealed');
        el.style.transform = 'translate3d(0, 30px, 0) scale(0.96)';
        el.style.opacity = '0';
      }
    });

    // 3. Kinetic Marquee Crawl: speed up crawl with scroll velocity
    if (crawlTracks.length && scrollVelocity > 1) {
      const boost = Math.min(3, 1 + scrollVelocity * 0.04);
      crawlTracks.forEach((track) => {
        track.style.animationDuration = `${Math.max(12, 35 / boost)}s`;
      });
    }

    isTicking = false;
  }

  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll, { passive: true });

  // Initial trigger after DOM loads
  window.addEventListener('DOMContentLoaded', () => {
    updateCinematics();
    setTimeout(updateCinematics, 100);
  });
  updateCinematics();
}
