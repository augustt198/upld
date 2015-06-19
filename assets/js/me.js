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

function sendPost(url, data, okCallback, errCallback) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", url, true);
    xhr.onload = function() {
        if (xhr.status == 200) {
            okCallback(xhr.responseText);
        } else {
            errCallback();
        }
    }
    xhr.send(data);
}

function favoriteUpload(id) {
    sendPost("/favorite", id, function() {
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
    sendPost("/delete", id, function(res) {
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

function getSelectedIds() {
    var checkboxes = getCheckboxes();
    var selected = [];
    for (var i = 0; i < checkboxes.length; i++) {
        if (checkboxes[i].checked) {
            selected.push(checkboxes[i].dataset.id);
        }
    }
    return selected;
}

function multiDelete() {
    var ids = getSelectedIds();

    var msg = "Are you sure you want to delete " + ids.length + " item";
    msg += ids.length == 1 ? "?" : "s?";

    if (!confirm(msg)) {
        return;
    }

    var deleteBtn = document.getElementById("delete-action");
    deleteBtn.innerHTML = "Deleting...";

    var data = ids.join(",");
    sendPost("/delete", data, function(res) {
        var removed = res.split(",");
        for (var i = 0; i < removed.length; i++) {
            document.getElementById("media-" + removed[i]).remove();
        }
        deleteBtn.innerHTML = "Delete";
    }, function() {
        deleteBtn.innerHTML = "Delete";
        alert("Failure during batch delete of IDs: " + data);
    });
}

// state:
// true = favorite multiple
// false = unfavorite multiple
function multiFavorite(state) {
    var ids = getSelectedIds();

    var action = state ? "favorite " : "unfavorite ";
    var msg = "Are you sure you want to " + action + ids.length + " item";
    msg += ids.length == 1 ? "?" : "s?";

    if (!confirm(msg)) {
        return;
    }

    var btn = document.getElementById((state ? "" : "un") + "fav-action");
    var originalText = btn.innerHTML;
    if (state) {
        btn.innerHTML = "Favoriting...";
    } else {
        btn.innerHTML = "Unfavoriting...";
    }

    var data = ids.join(",");
    var url = "/favorite?fav=" + (state ? "1" : "0");
    sendPost(url, data, function(res) {
        var removed = res.split(",");
        for (var i = 0; i < removed.length; i++) {
            var selector = "#media-" + removed[i] + " #menu-favorite-item";
            var elem = document.querySelector(selector);
            if (state) {
                elem.innerHTML = "Unfavorite";
            } else {
                elem.innerHTML = "Favorite";
            }
        }

        btn.innerHTML = originalText;        
    }, function() {
        btn.innerHTML = originalText;
        alert("Failure during batch (un)favorite of IDs: " + data);
    });
}
