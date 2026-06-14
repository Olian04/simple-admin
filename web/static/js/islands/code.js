// Code widget: mounts the Monaco editor (loaded from the configured CDN) and
// syncs its content back to a hidden textarea for form submission.
(function () {
  let loaded = false;
  let loading = false;
  const queue = [];

  function cdnBase() {
    return document.body.getAttribute("data-monaco-cdn");
  }

  function ensure(cb) {
    if (loaded) { cb(); return; }
    queue.push(cb);
    if (loading) return;
    loading = true;
    const base = cdnBase();
    const loader = document.createElement("script");
    loader.src = base + "/loader.js";
    loader.onload = function () {
      /* global require */
      window.require.config({ paths: { vs: base } });
      window.MonacoEnvironment = {
        getWorkerUrl: function () {
          const proxy =
            "self.MonacoEnvironment={baseUrl:'" + base + "/'};" +
            "importScripts('" + base + "/base/worker/workerMain.js');";
          return "data:text/javascript;charset=utf-8," + encodeURIComponent(proxy);
        },
      };
      window.require(["vs/editor/editor.main"], function () {
        loaded = true;
        queue.forEach(function (f) { f(); });
        queue.length = 0;
      });
    };
    document.head.appendChild(loader);
  }

  function init(root) {
    (root || document).querySelectorAll("[data-sa-code]").forEach(function (wrap) {
      if (wrap.dataset.saMounted) return;
      wrap.dataset.saMounted = "1";
      const ta = wrap.querySelector("textarea");
      const host = wrap.querySelector(".sa-code-editor");
      const lang = wrap.getAttribute("data-language") || "plaintext";
      ensure(function () {
        /* global monaco */
        const ed = window.monaco.editor.create(host, {
          value: ta.value,
          language: lang,
          automaticLayout: true,
          minimap: { enabled: false },
        });
        ed.onDidChangeModelContent(function () { ta.value = ed.getValue(); });
      });
    });
  }

  document.addEventListener("DOMContentLoaded", function () { init(document); });
  document.addEventListener("htmx:afterSwap", function (e) { init(e.target); });
})();
