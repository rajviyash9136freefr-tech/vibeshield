/**
 * Apple Cinematic Layer — Pure Two-Color Lighting & Motion
 *  1. Zero-latency custom mouse follower circle
 *  2. Full 3D card tilt [data-tilt] + specular edge spotlight + [data-depth]
 *  3. Ambient flashlight tracker across the document
 *  4. [data-countup] number roll-up on first scroll-into-view
 */

const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');

// ---- 1. Zero-Latency Custom Mouse Follower Circle -----------------------
if (finePointer.matches && typeof window !== 'undefined') {
  const dot = document.getElementById('cursor-dot');
  const ring = document.getElementById('cursor-ring');

  if (dot && ring) {
    let mouseX = -100;
    let mouseY = -100;
    let ringX = -100;
    let ringY = -100;
    let isHovering = false;
    let isVisible = false;

    // Immediate zero-latency update for the center dot
    window.addEventListener(
      'pointermove',
      (e) => {
        mouseX = e.clientX;
        mouseY = e.clientY;

        if (!isVisible) {
          dot.style.opacity = '1';
          ring.style.opacity = '1';
          ringX = mouseX;
          ringY = mouseY;
          isVisible = true;
        }

        // Hardware-accelerated 0ms direct positioning
        dot.style.transform = `translate3d(${mouseX}px, ${mouseY}px, 0)`;
      },
      { passive: true }
    );

    // Fluid RAF loop for the trailing magnetic ring
    function renderCursor() {
      // 0.32 lerp factor guarantees ultra-responsive, zero noticeable lag feel
      ringX += (mouseX - ringX) * 0.32;
      ringY += (mouseY - ringY) * 0.32;

      const scale = isHovering ? 1.55 : 1;
      ring.style.transform = `translate3d(${ringX}px, ${ringY}px, 0) scale(${scale})`;

      requestAnimationFrame(renderCursor);
    }
    requestAnimationFrame(renderCursor);

    // Interactive element detection (hover state expands the ring)
    const interactiveSelector =
      'a, button, [data-tilt], input, select, textarea, summary, [data-copy], .card';

    document.addEventListener(
      'pointerover',
      (e) => {
        const target = e.target as HTMLElement | null;
        if (target && target.closest(interactiveSelector)) {
          isHovering = true;
          ring.style.borderColor = 'rgba(255, 255, 255, 0.9)';
          ring.style.backgroundColor = 'rgba(255, 255, 255, 0.08)';
          ring.style.boxShadow = '0 0 16px rgba(255, 255, 255, 0.2)';
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
          ring.style.borderColor = 'rgba(255, 255, 255, 0.4)';
          ring.style.backgroundColor = 'transparent';
          ring.style.boxShadow = 'none';
        }
      },
      { passive: true }
    );

    // Fade out when pointer leaves browser window
    document.addEventListener('mouseleave', () => {
      dot.style.opacity = '0';
      ring.style.opacity = '0';
      isVisible = false;
    });
  }
}

// ---- 2. Full 3D Tilt + Specular Spotlight + Depth -----------------------
if (finePointer.matches) {
  document.querySelectorAll<HTMLElement>('[data-tilt]').forEach((card) => {
    const max = Number(card.dataset.tilt || 5);

    card.addEventListener('pointermove', (e) => {
      const r = card.getBoundingClientRect();
      const px = (e.clientX - r.left) / r.width;
      const py = (e.clientY - r.top) / r.height;

      // Dynamic 3D tilt calculation
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

