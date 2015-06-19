var activeDropdown = null;

function deactivateDropdown() {
    if (activeDropdown) {
        activeDropdown.setAttribute("style", "display: none");
        activeDropdown = null;
    }    
}

document.querySelector("html").addEventListener("click", function() {
    deactivateDropdown();
});

function activateUploadMenu(id) {
    var elem = document.getElementById("upload-menu-" + id);
    elem.setAttribute("style", "display: block");
    activeDropdown = elem;

    elem.addEventListener("click", function(e) {
        e.stopPropagation();
    });
}

function sendPost(url, okCallback, errCallback) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", url, true);
    xhr.onload = function() {
        if (xhr.status == 200) {
            okCallback();
        } else {
            errCallback();
        }
    }
    xhr.send();
}

function favoriteUpload(id) {
    sendPost("/favorite/" + id, function() {
        var elem = document.querySelector("#media-" + id + " #menu-favorite-item");
        if (elem.innerHTML.trim() == "Favorite") {
            elem.innerHTML = "Unfavorite";
        } else {
            elem.innerHTML = "Favorite";
        }
    }, function() {
        alert("Error occured while toggling favorite for " + id);
    });
}

function deleteUpload(id) {
    var elem = document.getElementById("menu-delete-item");
    elem.innerHTML = "Deleting...";
    sendPost("/delete/" + id, function() {
        document.getElementById("media-" + id).remove();
    }, function() {
        alert("Error occured while deleting " + id);
        elem.innerHTML = "Delete"
    });
}
