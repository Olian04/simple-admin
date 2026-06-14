// Tags widget: chip input backed by a comma-separated hidden field. No deps.
(function () {
  function values(wrap) {
    const hidden = wrap.querySelector("input[type=hidden]");
    return (hidden.value || "")
      .split(",")
      .map(function (s) { return s.trim(); })
      .filter(Boolean);
  }

  function sync(wrap, vals) {
    wrap.querySelector("input[type=hidden]").value = vals.join(",");
  }

  function render(wrap) {
    const list = wrap.querySelector(".sa-tags-list");
    list.innerHTML = "";
    values(wrap).forEach(function (v, i) {
      const tag = document.createElement("span");
      tag.className = "tag is-info";
      tag.textContent = v;
      const del = document.createElement("button");
      del.type = "button";
      del.className = "delete is-small";
      del.setAttribute("data-sa-tag-remove", String(i));
      tag.appendChild(del);
      list.appendChild(tag);
    });
  }

  function init(root) {
    (root || document).querySelectorAll("[data-sa-tags]").forEach(render);
  }

  document.addEventListener("keydown", function (e) {
    const inp = e.target.closest("[data-sa-tags-input]");
    if (!inp || e.key !== "Enter") return;
    e.preventDefault();
    const wrap = inp.closest("[data-sa-tags]");
    const v = inp.value.trim();
    if (v) {
      const arr = values(wrap);
      if (arr.indexOf(v) < 0) {
        arr.push(v);
        sync(wrap, arr);
        render(wrap);
      }
      inp.value = "";
    }
  });

  document.addEventListener("click", function (e) {
    const del = e.target.closest("[data-sa-tag-remove]");
    if (!del) return;
    const wrap = del.closest("[data-sa-tags]");
    const arr = values(wrap);
    arr.splice(parseInt(del.getAttribute("data-sa-tag-remove"), 10), 1);
    sync(wrap, arr);
    render(wrap);
  });

  document.addEventListener("DOMContentLoaded", function () { init(document); });
  document.addEventListener("htmx:afterSwap", function (e) { init(e.target); });
})();
