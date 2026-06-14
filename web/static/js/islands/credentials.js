// Browser-side credential store + HTMX passthrough injector.
//
// Per-backend secrets live in localStorage keyed "simple-admin:cred:{backendId}".
// They are attached to outgoing HTMX requests for the matching backend as the
// X-Sa-Credential header; the server extracts it and passes it through to the
// upstream backend. Secrets never persist server-side.
(function () {
  const PREFIX = "simple-admin:cred:";
  const keyFor = (id) => PREFIX + id;

  // /ui/{backend}/... -> backend id
  function backendFromPath(path) {
    const m = /^\/ui\/([^/]+)\//.exec(path || "");
    return m ? m[1] : null;
  }

  // Inject the active backend's credential into HTMX requests.
  document.addEventListener("htmx:configRequest", function (evt) {
    const id = backendFromPath(evt.detail.path);
    if (!id) return;
    const cred = localStorage.getItem(keyFor(id));
    if (cred) {
      evt.detail.headers["X-Sa-Credential"] = cred;
      evt.detail.headers["X-Sa-Backend"] = id;
    }
  });

  // Credentials page: populate inputs from localStorage.
  function populate() {
    document.querySelectorAll("[data-sa-cred-input]").forEach(function (input) {
      const v = localStorage.getItem(keyFor(input.getAttribute("data-backend")));
      if (v) input.value = v;
    });
  }

  function flash(btn, text) {
    const original = btn.textContent;
    btn.textContent = text;
    setTimeout(function () {
      btn.textContent = original;
    }, 1200);
  }

  document.addEventListener("click", function (e) {
    const save = e.target.closest("[data-sa-cred-save]");
    if (save) {
      const id = save.getAttribute("data-backend");
      const input = document.querySelector('[data-sa-cred-input][data-backend="' + id + '"]');
      if (input) {
        localStorage.setItem(keyFor(id), input.value);
        flash(save, "Saved");
      }
      return;
    }
    const clear = e.target.closest("[data-sa-cred-clear]");
    if (clear) {
      const id = clear.getAttribute("data-backend");
      localStorage.removeItem(keyFor(id));
      const input = document.querySelector('[data-sa-cred-input][data-backend="' + id + '"]');
      if (input) input.value = "";
      flash(clear, "Cleared");
    }
  });

  document.addEventListener("DOMContentLoaded", populate);
  document.addEventListener("htmx:afterSwap", populate);
})();
