/**
 * Cinematic layer — built on UIUX §2.4 discipline (one accent, bordered
 * elevation, reduced-motion disables everything). Three behaviors:
 *  1. [data-tilt]   3D pointer tilt + cursor spotlight (lg + fine pointer only)
 *  2. [data-countup] number roll-up on first scroll-into-view (once)
 * Scroll-driven parallax/progress/rail live in global.css (@supports
 * animation-timeline) — pure CSS where the browser can do it, no JS needed.
 */
const reduced = document.documentElement.dataset.reducedMotion === '1';
const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');
const wide = window.matchMedia('(min-width: 1024px)');

// ---- 3D tilt + spotlight -------------------------------------------------
if (!reduced && finePointer.matches && wide.matches) {
  document.querySelectorAll<HTMLElement>('[data-tilt]').forEach((card) => {
    const max = Number(card.dataset.tilt || 6);
    card.addEventListener('pointermove', (e) => {
      const r = card.getBoundingClientRect();
      const px = (e.clientX - r.left) / r.width;
      const py = (e.clientY - r.top) / r.height;
      card.style.setProperty('--tilt-x', `${((py - 0.5) * -2 * max).toFixed(2)}deg`);
      card.style.setProperty('--tilt-y', `${((px - 0.5) * 2 * max).toFixed(2)}deg`);
      card.style.setProperty('--mx', `${(px * 100).toFixed(1)}%`);
      card.style.setProperty('--my', `${(py * 100).toFixed(1)}%`);
    });
    card.addEventListener('pointerleave', () => {
      card.style.setProperty('--tilt-x', '0deg');
      card.style.setProperty('--tilt-y', '0deg');
    });
  });
}

// ---- count-up on first reveal --------------------------------------------
const counters = document.querySelectorAll<HTMLElement>('[data-countup]');
if (!reduced && counters.length && 'IntersectionObserver' in window) {
  const io = new IntersectionObserver(
    (entries) => {
      for (const en of entries) {
        if (!en.isIntersecting) continue;
        const el = en.target as HTMLElement;
        io.unobserve(el);
        const raw = el.dataset.countup || '0';
        const target = parseFloat(raw);
        const decimals = (raw.split('.')[1] || '').length;
        const suffix = el.dataset.suffix || '';
        const t0 = performance.now();
        const dur = 900;
        const tick = (t: number) => {
          const k = Math.min(1, (t - t0) / dur);
          const eased = 1 - Math.pow(1 - k, 3); // matches --ease-soft feel
          el.textContent = (target * eased).toFixed(decimals) + suffix;
          if (k < 1) requestAnimationFrame(tick);
        };
        requestAnimationFrame(tick);
      }
    },
    { threshold: 0.5 },
  );
  counters.forEach((el) => io.observe(el));
}
