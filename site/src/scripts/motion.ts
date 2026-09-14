/**
 * Apple Cinematic Layer — Pure Two-Color Lighting & Motion
 *  1. [data-tilt]   3D pointer tilt + specular edge spotlight
 *  2. [data-countup] number roll-up on first scroll-into-view
 *  3. Ambient flashlight tracker across the document
 */
const reduced = document.documentElement.dataset.reducedMotion === '1';
const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');
const wide = window.matchMedia('(min-width: 1024px)');

// ---- Ambient flashlight tracking -----------------------------------------
if (!reduced && finePointer.matches) {
  let targetX = window.innerWidth / 2;
  let targetY = window.innerHeight * 0.3;
  let currentX = targetX;
  let currentY = targetY;
  let rafId: number | null = null;

  window.addEventListener('pointermove', (e) => {
    targetX = e.clientX;
    targetY = e.clientY;
    if (!rafId) {
      rafId = requestAnimationFrame(updateGlow);
    }
  }, { passive: true });

  function updateGlow() {
    currentX += (targetX - currentX) * 0.15;
    currentY += (targetY - currentY) * 0.15;

    document.documentElement.style.setProperty('--mouse-x', `${currentX.toFixed(1)}px`);
    document.documentElement.style.setProperty('--mouse-y', `${currentY.toFixed(1)}px`);

    if (Math.abs(targetX - currentX) > 0.5 || Math.abs(targetY - currentY) > 0.5) {
      rafId = requestAnimationFrame(updateGlow);
    } else {
      rafId = null;
    }
  }
}

// ---- 3D tilt + specular spotlight ---------------------------------------
if (!reduced && finePointer.matches && wide.matches) {
  document.querySelectorAll<HTMLElement>('[data-tilt]').forEach((card) => {
    const max = Number(card.dataset.tilt || 5);
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

// ---- Count-up on first reveal --------------------------------------------
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
        const dur = 1000;
        const tick = (t: number) => {
          const k = Math.min(1, (t - t0) / dur);
          const eased = 1 - Math.pow(1 - k, 3); // Apple smooth cubic ease-out
          el.textContent = (target * eased).toFixed(decimals) + suffix;
          if (k < 1) requestAnimationFrame(tick);
        };
        requestAnimationFrame(tick);
      }
    },
    { threshold: 0.4 },
  );
  counters.forEach((el) => io.observe(el));
}
