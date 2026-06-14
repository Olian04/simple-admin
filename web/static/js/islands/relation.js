// Relation picker island: populates <select data-sa-relation> with options
// fetched from the admin's /options endpoint for the target resource.
(function () {
  const PREFIX = "simple-admin:cred:";

  function backendFromURL(url) {
    const m = /^\/ui\/([^/]+)\//.exec(url || "");
    return m ? m[1] : null;
  }

  // Attach the passthrough credential for gated backends (plain fetch bypasses
  // the HTMX configRequest injector).
  function credHeaders(url) {
    const id = backendFromURL(url);
    const headers = {};
    if (id) {
      const cred = localStorage.getItem(PREFIX + id);
      if (cred) {
        headers["X-Sa-Credential"] = cred;
        headers["X-Sa-Backend"] = id;
      }
    }
    return headers;
  }

  function init(root) {
    (root || document).querySelectorAll("select[data-sa-relation]").forEach(function (sel) {
      if (sel.dataset.saLoaded) return;
      sel.dataset.saLoaded = "1";
      const url = sel.getAttribute("data-options-url");
      const current = sel.getAttribute("data-current") || "";
      if (!url) return;
      fetch(url, { headers: credHeaders(url) })
        .then(function (r) { return r.ok ? r.json() : []; })
        .then(function (opts) {
          Array.from(sel.options).forEach(function (o) {
            if (o.value !== "") o.remove();
          });
          (opts || []).forEach(function (o) {
            const el = document.createElement("option");
            el.value = o.value;
            el.textContent = o.label;
            if (o.value === current) el.selected = true;
            sel.appendChild(el);
          });
        })
        .catch(function () {});
    });
  }

  document.addEventListener("DOMContentLoaded", function () { init(document); });
  document.addEventListener("htmx:afterSwap", function (e) { init(e.target); });
})();
