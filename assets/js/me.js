var activeDropdown = null;

document.querySelector("html").addEventListener("click", function() {
    if (activeDropdown) {
        activeDropdown.setAttribute("style", "display: none");
        activeDropdown = null;
    }
});

function activateUploadMenu(id) {
    var elem = document.getElementById("upload-menu-" + id);
    elem.setAttribute("style", "display: block");
    activeDropdown = elem;

    elem.addEventListener("click", function(e) {
        e.stopPropagation();
    });
}
