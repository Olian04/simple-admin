// simple-admin small UI helpers (CSP-safe; no inline handlers).

// Delegated modal close: any element with [data-sa-close] removes its modal.
// Flash close: [data-sa-flash-close] removes its flash notification.
document.addEventListener("click", function (e) {
  const closer = e.target.closest("[data-sa-close]");
  if (closer) {
    const modal = closer.closest(".modal");
    if (modal) modal.remove();
    return;
  }
  const flashCloser = e.target.closest("[data-sa-flash-close]");
  if (flashCloser) {
    const flash = flashCloser.closest("[data-sa-flash]");
    if (flash) flash.remove();
  }
});

// Auto-dismiss flash notifications after a few seconds.
function saAutoDismissFlash(root) {
  (root || document).querySelectorAll("[data-sa-flash]").forEach(function (el) {
    if (el.dataset.saTimer) return;
    el.dataset.saTimer = "1";
    setTimeout(function () {
      el.remove();
    }, 4000);
  });
}
document.addEventListener("DOMContentLoaded", function () {
  saAutoDismissFlash(document);
});
document.addEventListener("htmx:afterSwap", function (e) {
  saAutoDismissFlash(e.target);
});
