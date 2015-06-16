if (document.getElementById("me-page-menu") != null) {
    var children = document.getElementById("me-page-menu").children;
    var path = window.location.pathname;
    for (var i = 0; i < children.length; i++) {
        var e = children[i];
        if (e.tagName == "A" && e.getAttribute("href") == path) {
            var cls = e.getAttribute("class") + " selected";
            e.setAttribute("class", cls);
            break;
        }
    }
}
