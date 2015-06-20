// http://blog.teamtreehouse.com/uploading-files-ajax
var form = document.getElementById("upload-form");
var fileSelector = document.getElementById("file-selector");
var uploadButton = document.getElementById("upload-button");

// 20MB
var MAX_SIZE = 20 * 1000000;

form.onsubmit = function(e) {
    e.preventDefault();

    var file = fileSelector.files[0];
    if (file.size > MAX_SIZE) {
        alert("File above maximum size: 20MB");
        return;
    }
    
    var formData = new FormData();
    formData.append("upload", file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/upload", true);
    xhr.upload.addEventListener("progress", updateProgress, false);

    uploadButton.setAttribute("value", "Uploading 0%");
    var container = document.getElementById("progress-bar-container");
    clearChildren(container);
    var bar = document.createElement("div");
    bar.setAttribute("class", "progress-bar");
    var fill = document.createElement("div");
    fill.setAttribute("class", "progress-bar-fill");
    fill.setAttribute("style", "width: 0")
    bar.appendChild(fill);
    container.appendChild(bar);

    function updateProgress(event) {
        if (event.lengthComputable) {
            var raw = event.loaded / event.total
            var width = Math.round(raw * 298);
            var percent = Math.round(raw * 100);
            
            fill.setAttribute("style", "width: " + width + "px");
            if (percent == 100) {
                uploadButton.setAttribute("value", "Processing...");
            } else {
                uploadButton.setAttribute("value", "Uploading " + percent + "%");
            }
        }
    }

    xhr.onload = function() {
        if (xhr.status == 200) {
            window.location = "/view/" + xhr.responseText;
        } else {
            uploadButton.setAttribute("class", "btn btn-danger");
            uploadButton.setAttribute("value", "Error")
            console.log("Error: " + xhr.status);
        }
    }

    xhr.send(formData);
}

function addFiles() {
    var children = document.getElementById("hidden-inputs-wrapper").children;
    var last = children[children.length - 1];

    last.click();
}

function handleSelectedFiles(elem) {
    var elemParent = elem.parentNode;

    var node = document.createElement("INPUT");
    node.setAttribute("type", "file");
    node.setAttribute("multiple", "");
    node.onchange = function() { handleSelectedFiles(node) };
    node.dataset.id = parseInt(elem.dataset.id) + 1;
    elemParent.appendChild(node);

    var files = elem.files;
    for (var i = 0; i < files.length; i++) {
        uploadFile(files[i]);
    }
}

function uploadFile(file) {
    var list = document.getElementById("uploads-list");

    var node = document.getElementById("sample-upload-entry").cloneNode(true);
    node.setAttribute("id", "last-upload-entry");
    node.setAttribute("style", "");

    var textSpan = node.children[0];
    var progressSpan = node.children[1];
    var progressBar = node.children[2]
    
    textSpan.appendChild(document.createTextNode(file.name));

    list.appendChild(node);

    var data = new FormData();
    data.append("upload", file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/upload", true);
    xhr.upload.addEventListener("progress", updateProgress, false);

    function updateProgress(event) {
        if (event.lengthComputable) {
            var raw = event.loaded / event.total;
            var percent = Math.round(raw * 100);

            if (percent == 100) {
                progressSpan.innerHTML = "Processing...";
                var prevClass = progressBar.getAttribute("class");
                progressBar.setAttribute("class", prevClass + " striped-progress-bg");

            } else {
                progressSpan.innerHTML = percent + "%"
                progressBar.setAttribute("style", "width: " + (raw * 100) + "%");    
            }
            
        }
    }

    xhr.onload = function() {
        if (xhr.status == 200) {
            progressBar.setAttribute("class", "progress-bg progress-complete-bg");

            var linkNode = document.createElement("A");
            linkNode.setAttribute("href", "/view/" + xhr.responseText);
            linkNode.setAttribute("target", "_blank");
            linkNode.appendChild(document.createTextNode(file.name));
            clearChildren(textSpan);
            textSpan.appendChild(linkNode);

            progressSpan.innerHTML = "Complete";
        } else {
            progressBar.setAttribute("class", "progress-bg progress-failure-bg");
            progressSpan.innerHTML = "Failed";
        }
    }

    xhr.send(data);
}
