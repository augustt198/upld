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

var masterCheckbox = document.getElementById("select-all-checkbox");
function getCheckboxes() {
    return document.querySelectorAll(".media-item .media-corner input");
}

function areAllSelected() {
    var checkboxes = getCheckboxes();
    for (var i = 0; i < checkboxes.length; i++) {
        if (!checkboxes[i].checked) {
            return false;
        }
    }
    return true;
}

function countChecked() {
    var checkboxes = getCheckboxes();
    var count = 0;
    for (var i = 0; i < checkboxes.length; i++) {
        if (checkboxes[i].checked)
            count++;
    }
    return count;
}

function actionsVisible(visible) {
    var elem = document.getElementById("action-container");
    if (visible) {
        elem.setAttribute("style", "display: inline-block");
    } else {
        elem.setAttribute("style", "display: none");
    }
}

var allCheckboxes = getCheckboxes();
for (var i = 0; i < allCheckboxes.length; i++) {
    allCheckboxes[i].onchange = function() {
        var all = areAllSelected();
        masterCheckbox.checked = all;
        var count = countChecked();
        if (all || count > 0) {
            actionsVisible(true)
        } else {
            actionsVisible(false);
        }
    }
}


function selectAllBtn() {
    var newState = !masterCheckbox.checked;
    masterCheckbox.checked = newState;

    selectAllCheckbox();
}

function selectAllCheckbox() {
    var state = masterCheckbox.checked;
    var checkboxes = getCheckboxes();

    for (var i = 0; i < checkboxes.length; i++) {
        checkboxes[i].checked = state;
    }

    actionsVisible(state);
}
