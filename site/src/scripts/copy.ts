/**
 * Copy-to-clipboard for every [data-copy] button on the page (~10 lines,
 * UIUX §5 code-block component law). Progressive: button hides without JS.
 */
document.querySelectorAll<HTMLButtonElement>('[data-copy]').forEach((btn) => {
  if (!navigator.clipboard) {
    btn.classList.add('hidden');
    return;
  }
  btn.addEventListener('click', async () => {
    const code = btn.closest('div')?.querySelector<HTMLElement>('[data-code]');
    if (!code) return;
    await navigator.clipboard.writeText(code.innerText);
    const prev = btn.textContent;
    btn.textContent = 'Copied';
    setTimeout(() => (btn.textContent = prev), 1500);
  });
});
