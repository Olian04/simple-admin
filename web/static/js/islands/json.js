// JSON widget: a "Format" button that pretty-prints / validates the textarea.
(function () {
  document.addEventListener("click", function (e) {
    const btn = e.target.closest("[data-sa-json-format]");
    if (!btn) return;
    const wrap = btn.closest("[data-sa-json]");
    if (!wrap) return;
    const ta = wrap.querySelector("textarea");
    const err = wrap.querySelector("[data-sa-json-error]");
    try {
      ta.value = JSON.stringify(JSON.parse(ta.value), null, 2);
      if (err) {
        err.textContent = "";
        err.classList.add("is-hidden");
      }
    } catch (ex) {
      if (err) {
        err.textContent = "Invalid JSON: " + ex.message;
        err.classList.remove("is-hidden");
      }
    }
  });
})();
