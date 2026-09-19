// Bidirectional countup runner for [data-countup] elements
const counters = document.querySelectorAll<HTMLElement>('[data-countup]');
if (counters.length && typeof window !== 'undefined' && 'IntersectionObserver' in window) {
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

    const duration = 800;
    const startTime = performance.now();

    const prevRaf = Number(el.dataset.countupRaf || 0);
    if (prevRaf) cancelAnimationFrame(prevRaf);

    function step(now: number) {
      const progress = Math.min(1, (now - startTime) / duration);
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

  const counterObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          runCountUp(entry.target as HTMLElement);
        }
      }
    },
    { threshold: 0.2 }
  );

  counters.forEach((el) => counterObserver.observe(el));
}
