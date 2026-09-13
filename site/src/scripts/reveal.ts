// Scroll-into-view entrances — UIUX §2.4: fade + 8px rise, once, staggered
// 60ms. Reduced motion: revealed immediately (handled in CSS too).
const els = document.querySelectorAll<HTMLElement>('[data-reveal]');
if (els.length) {
  if (document.documentElement.dataset.reducedMotion || !('IntersectionObserver' in window)) {
    els.forEach((el) => el.classList.add('revealed'));
  } else {
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            (e.target as HTMLElement).classList.add('revealed');
            io.unobserve(e.target);
          }
        }
      },
      { threshold: 0.15 },
    );
    els.forEach((el, i) => {
      el.style.setProperty('--reveal-delay', `${(i % 6) * 60}ms`);
      io.observe(el);
    });
  }
}
