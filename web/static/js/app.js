// simple-admin small UI helpers (CSP-safe; no inline handlers).

// Delegated modal close: any element with [data-sa-close] removes its modal.
document.addEventListener("click", function (e) {
  const closer = e.target.closest("[data-sa-close]");
  if (!closer) return;
  const modal = closer.closest(".modal");
  if (modal) modal.remove();
});
