/**
 * Medium Apple Cinematic Layer — Pure Two-Color Lighting & Motion
 *  1. Single Custom Circle Cursor (black center, feathered white border)
 *  2. Balanced medium 3D card tilt [data-tilt] + subtle specular spotlight
 *  3. Subtle ambient flashlight tracker across the document
 *  4. [data-countup] number roll-up on first scroll-into-view
 */

const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');

// ---- 1. Single Circle Cursor (Center Black, Feathered White Border) -----
if (finePointer.matches && typeof window !== 'undefined') {
  const cursor = document.getElementById('cursor-circle');

  if (cursor) {
    let mouseX = -100;
    let mouseY = -100;
    let currentX = -100;
    let currentY = -100;
    let isHovering = false;
    let isVisible = false;

    window.addEventListener(
      'pointermove',
      (e) => {
        mouseX = e.clientX;
        mouseY = e.clientY;

        if (!isVisible) {
          cursor.style.opacity = '1';
          currentX = mouseX;
          currentY = mouseY;
          isVisible = true;
        }
      },
      { passive: true }
    );

    // Highly responsive, smooth RAF loop (0.45 lerp factor: zero noticeable lag, velvety glide)
    function renderCursor() {
      currentX += (mouseX - currentX) * 0.45;
      currentY += (mouseY - currentY) * 0.45;

      const scale = isHovering ? 1.25 : 1;
      cursor.style.transform = `translate3d(${currentX.toFixed(1)}px, ${currentY.toFixed(1)}px, 0) scale(${scale})`;

      requestAnimationFrame(renderCursor);
    }
    requestAnimationFrame(renderCursor);

    // Interactive element detection (hover state gently expands feathered white circle)
    const interactiveSelector =
      'a, button, [data-tilt], input, select, textarea, summary, [data-copy], .card';

    document.addEventListener(
      'pointerover',
      (e) => {
        const target = e.target as HTMLElement | null;
        if (target && target.closest(interactiveSelector)) {
          isHovering = true;
          cursor.classList.add('cursor-hover');
        }
      },
      { passive: true }
    );

    document.addEventListener(
      'pointerout',
      (e) => {
        const target = e.target as HTMLElement | null;
        if (target && target.closest(interactiveSelector)) {
          isHovering = false;
          cursor.classList.remove('cursor-hover');
        }
      },
      { passive: true }
    );

    // Fade out when pointer leaves browser window
    document.addEventListener('mouseleave', () => {
      cursor.style.opacity = '0';
      isVisible = false;
    });
  }
}

// ---- 2. Balanced Medium 3D Card Tilt + Specular Edge -------------------
if (finePointer.matches) {
  document.querySelectorAll<HTMLElement>('[data-tilt]').forEach((card) => {
    // Restrained medium tilt angle (max 1.5 - 1.8 degrees for subtle tactile response)
    const configured = Number(card.dataset.tilt || 2);
    const max = Math.min(1.8, Math.max(1.0, configured * 0.4));

    card.addEventListener('pointermove', (e) => {
      const r = card.getBoundingClientRect();
      const px = (e.clientX - r.left) / r.width;
      const py = (e.clientY - r.top) / r.height;

      // Subtle, tactile 3D tilt calculation
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

// ---- 4. Count-up on First Reveal -----------------------------------------
const counters = document.querySelectorAll<HTMLElement>('[data-countup]');
if (counters.length && 'IntersectionObserver' in window) {
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
    { threshold: 0.3 }
  );
  counters.forEach((el) => io.observe(el));
}

