/**
 * Medium Apple Cinematic Layer — Pure Two-Color Lighting & Motion
 *  1. Bespoke Aero Reticle Cursor (Laser Precision Core + Floating Magnetic Aura)
 *  2. Balanced medium 3D card tilt [data-tilt] + subtle specular spotlight
 *  3. Subtle ambient flashlight tracker across the document
 *  4. Bidirectional [data-countup] number roll-up engine (scroll down & scroll up)
 */

const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');

// ---- 1. Bespoke Aero Reticle Cursor System ------------------------------
if (finePointer.matches && typeof window !== 'undefined') {
  const dot = document.getElementById('cursor-dot');
  const aura = document.getElementById('cursor-aura');

  if (dot && aura) {
    let mouseX = -100;
    let mouseY = -100;
    let auraX = -100;
    let auraY = -100;
    let isHovering = false;
    let isPressed = false;
    let isVisible = false;
    let dotScale = 1;

    // Instant zero-latency direct hardware tracking for the precision core dot
    window.addEventListener(
      'pointermove',
      (e) => {
        mouseX = e.clientX;
        mouseY = e.clientY;

        if (!isVisible) {
          dot.style.opacity = '1';
          aura.style.opacity = '1';
          auraX = mouseX;
          auraY = mouseY;
          isVisible = true;
        }

        dot.style.transform = `translate3d(${mouseX}px, ${mouseY}px, 0) scale(${dotScale})`;
      },
      { passive: true }
    );

    // Fluid spring-lerp loop for the floating magnetic aura ring
    function renderAeroCursor() {
      auraX += (mouseX - auraX) * 0.28;
      auraY += (mouseY - auraY) * 0.28;

      const auraScale = isPressed ? 0.8 : isHovering ? 1.45 : 1;
      aura.style.transform = `translate3d(${auraX.toFixed(1)}px, ${auraY.toFixed(1)}px, 0) scale(${auraScale})`;

      requestAnimationFrame(renderAeroCursor);
    }
    requestAnimationFrame(renderAeroCursor);

    // Interactive target detection (links, buttons, interactive cards, inputs)
    const interactiveSelector =
      'a, button, [data-tilt], input, select, textarea, summary, [data-copy], .card, [role="button"]';
    const textSelector = 'p, code, pre, h1, h2, h3, blockquote';

    document.addEventListener(
      'pointerover',
      (e) => {
        const target = e.target as HTMLElement | null;
        if (!target) return;

        if (target.closest(interactiveSelector)) {
          isHovering = true;
          dotScale = 1.35;
          aura.classList.add('cursor-hover');
          aura.classList.remove('cursor-text');
          dot.style.transform = `translate3d(${mouseX}px, ${mouseY}px, 0) scale(${dotScale})`;
        } else if (target.closest(textSelector)) {
          aura.classList.add('cursor-text');
        }
      },
      { passive: true }
    );

    document.addEventListener(
      'pointerout',
      (e) => {
        const target = e.target as HTMLElement | null;
        if (!target) return;

        if (target.closest(interactiveSelector)) {
          isHovering = false;
          dotScale = 1;
          aura.classList.remove('cursor-hover');
          dot.style.transform = `translate3d(${mouseX}px, ${mouseY}px, 0) scale(${dotScale})`;
        }
        if (target.closest(textSelector)) {
          aura.classList.remove('cursor-text');
        }
      },
      { passive: true }
    );

    // Tactile click press & release
    window.addEventListener(
      'pointerdown',
      () => {
        isPressed = true;
        aura.classList.add('cursor-active');
      },
      { passive: true }
    );

    window.addEventListener(
      'pointerup',
      () => {
        isPressed = false;
        aura.classList.remove('cursor-active');
      },
      { passive: true }
    );

    // Fade out smoothly when mouse leaves the browser viewport
    document.addEventListener('mouseleave', () => {
      dot.style.opacity = '0';
      aura.style.opacity = '0';
      isVisible = false;
    });

    document.addEventListener('mouseenter', () => {
      dot.style.opacity = '1';
      aura.style.opacity = '1';
      isVisible = true;
    });
  }
}

// ---- 2. Balanced Medium 3D Card Tilt + Specular Edge -------------------
if (finePointer.matches) {
  document.querySelectorAll<HTMLElement>('[data-tilt]').forEach((card) => {
    const configured = Number(card.dataset.tilt || 2);
    const max = Math.min(1.8, Math.max(1.0, configured * 0.4));

    card.addEventListener('pointermove', (e) => {
      const r = card.getBoundingClientRect();
      const px = (e.clientX - r.left) / r.width;
      const py = (e.clientY - r.top) / r.height;

      const rotX = ((py - 0.5) * -2 * max).toFixed(2);
      const rotY = ((px - 0.5) * 2 * max).toFixed(2);

      card.style.setProperty('--tilt-x', `${rotX}deg`);
      card.style.setProperty('--tilt-y', `${rotY}deg`);
      card.style.setProperty('--mx', `${(px * 100).toFixed(1)}%`);
      card.style.setProperty('--my', `${(py * 100).toFixed(1)}%`);
    });

    card.addEventListener('pointerleave', () => {
      card.style.setProperty('--tilt-x', '0deg');
      card.style.setProperty('--tilt-y', '0deg');
    });
  });
}

// ---- 3. Ambient Flashlight Tracking --------------------------------------
if (finePointer.matches) {
  let targetX = window.innerWidth / 2;
  let targetY = window.innerHeight * 0.3;
  let currentX = targetX;
  let currentY = targetY;
  let rafId: number | null = null;

  window.addEventListener(
    'pointermove',
    (e) => {
      targetX = e.clientX;
      targetY = e.clientY;
      if (!rafId) {
        rafId = requestAnimationFrame(updateGlow);
      }
    },
    { passive: true }
  );

  function updateGlow() {
    currentX += (targetX - currentX) * 0.18;
    currentY += (targetY - currentY) * 0.18;

    document.documentElement.style.setProperty('--mouse-x', `${currentX.toFixed(1)}px`);
    document.documentElement.style.setProperty('--mouse-y', `${currentY.toFixed(1)}px`);

    if (Math.abs(targetX - currentX) > 0.5 || Math.abs(targetY - currentY) > 0.5) {
      rafId = requestAnimationFrame(updateGlow);
    } else {
      rafId = null;
    }
  }
}

// ---- 4. Reliable Bidirectional Count-Up Engine (Scroll Down & Scroll Up) -
const counters = document.querySelectorAll<HTMLElement>('[data-countup]');
if (counters.length && 'IntersectionObserver' in window) {
  const isReduced = document.documentElement.dataset.reducedMotion === '1';

  function runCountUp(el: HTMLElement) {
    const raw = el.dataset.countup;
    if (!raw) return;
    const target = parseFloat(raw);
    const decimals = (raw.split('.')[1] || '').length;
    const suffix = el.dataset.suffix || '';

    if (isReduced) {
      el.textContent = target.toFixed(decimals) + suffix;
      return;
    }

    const duration = 1100;
    const startTime = performance.now();

    // Cancel any previous animation frame
    const prevRaf = Number(el.dataset.countupRaf || 0);
    if (prevRaf) cancelAnimationFrame(prevRaf);

    function step(now: number) {
      const progress = Math.min(1, (now - startTime) / duration);
      // Apple smooth cubic ease-out
      const eased = 1 - Math.pow(1 - progress, 3);
      const current = target * eased;
      el.textContent = current.toFixed(decimals) + suffix;

      if (progress < 1) {
        const nextId = requestAnimationFrame(step);
        el.dataset.countupRaf = String(nextId);
      } else {
        el.textContent = target.toFixed(decimals) + suffix;
        delete el.dataset.countupRaf;
      }
    }

    const rafId = requestAnimationFrame(step);
    el.dataset.countupRaf = String(rafId);
  }

  function resetCountUp(el: HTMLElement) {
    const raw = el.dataset.countup;
    if (!raw) return;
    const decimals = (raw.split('.')[1] || '').length;
    const suffix = el.dataset.suffix || '';

    const prevRaf = Number(el.dataset.countupRaf || 0);
    if (prevRaf) {
      cancelAnimationFrame(prevRaf);
      delete el.dataset.countupRaf;
    }
    el.textContent = (0).toFixed(decimals) + suffix;
  }

  // Pre-initialize counters to zero so when user scrolls down they see them count up!
  if (!isReduced) {
    counters.forEach((el) => resetCountUp(el));
  }

  const counterObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const el = entry.target as HTMLElement;
        if (entry.isIntersecting) {
          runCountUp(el);
        } else {
          // Reset when scrolled out of view so scrolling back (down or up) triggers the count-up
          if (!isReduced) {
            resetCountUp(el);
          }
        }
      }
    },
    {
      threshold: 0.15,
      rootMargin: '0px 0px -20px 0px',
    }
  );

  counters.forEach((el) => counterObserver.observe(el));
}

