// Markdown widget: mounts EasyMDE (loaded from the configured CDN) on the
// textarea. EasyMDE keeps the underlying textarea value in sync for submit.
(function () {
  let loaded = false;
  let loading = false;
  const queue = [];

  function cdnBase() {
    return document.body.getAttribute("data-easymde-cdn");
  }

  function ensure(cb) {
    if (loaded) { cb(); return; }
    queue.push(cb);
    if (loading) return;
    loading = true;
    const base = cdnBase();
    const css = document.createElement("link");
    css.rel = "stylesheet";
    css.href = base + "/easymde.min.css";
    document.head.appendChild(css);
    const s = document.createElement("script");
    s.src = base + "/easymde.min.js";
    s.onload = function () {
      loaded = true;
      queue.forEach(function (f) { f(); });
      queue.length = 0;
    };
    document.head.appendChild(s);
  }

  function init(root) {
    (root || document).querySelectorAll("textarea[data-sa-markdown]").forEach(function (ta) {
      if (ta.dataset.saMounted) return;
      ta.dataset.saMounted = "1";
      ensure(function () {
        /* global EasyMDE */
        new EasyMDE({ element: ta });
      });
    });
  }

  document.addEventListener("DOMContentLoaded", function () { init(document); });
  document.addEventListener("htmx:afterSwap", function (e) { init(e.target); });
})();
