// http://blog.teamtreehouse.com/uploading-files-ajax

// 20MB
var MAX_SIZE = 20 * 1000000;

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

    if (files.length > 50) {
        alert("Sorry, you can't upload more than 50 files at once");
    } else {
        for (var i = 0; i < files.length; i++) {
            uploadFile(files[i]);
        }
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

    if (file.size > MAX_SIZE) {
        progressBar.setAttribute("class", "progress-bg progress-failure-bg");
        progressSpan.innerHTML = "Failed (file >20MB)";
        return;
    }

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

var htmlNode = document.body.parentNode;
htmlNode.ondragover = dragEnter;
htmlNode.ondragleave = dragLeave;
htmlNode.ondrop = onDrop;

var isDragging = false;
var titleElem = document.getElementById("upload-list-title");

function dragEnter(e) {
    e.stopPropagation();
    e.preventDefault();

    if (!isDragging) {
        titleElem.innerHTML = "Release to upload";
        isDragging = true;
    }

}

function dragLeave(e) {
    e.stopPropagation();
    e.preventDefault();
    
    isDragging = false;
    titleElem.innerHTML = "Upload files";
}

function onDrop(e) {
    e.stopPropagation();
    e.preventDefault();

    isDragging = false;
    titleElem.innerHTML = "Upload files";

    files = e.dataTransfer.files;
    if (files.length > 50) {
        alert("Sorry, you can't upload more than 50 files at once");
    } else {
        for (var i = 0; i < files.length; i++) {
            uploadFile(files[i]);
        }        
    }
}
