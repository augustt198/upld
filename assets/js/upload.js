// http://blog.teamtreehouse.com/uploading-files-ajax

// 20MB
var MAX_SIZE = 20 * 1000000;

var queue = [];
var uploadInProgress = false;

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

    if (!uploadInProgress)
        uploadFromQueue();
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

    queue.push([file, textSpan, progressSpan, progressBar]);
}

function uploadFromQueue() {
    if (queue.length > 0) {
        uploadInProgress = true;
        group = queue[0];


        startUpload(group[0], group[1], group[2], group[3])
    } else {
        uploadInProgress = false;
    }
}

function startUpload(file, textSpan, progressSpan, progressBar) {
    var form = new FormData();
    form.append("name", file.name);
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/upload_start");

    xhr.onload = function() {
        if (xhr.status == 200) {
            json = JSON.parse(xhr.responseText);
            uploadItem(file, textSpan, progressSpan, progressBar, json);
        } else {
            progressBar.setAttribute("class", "progress-bg progress-failure-bg");
            progressSpan.innerHTML = "Failed to authorize upload";
            queue.splice(0, 1);
            uploadFromQueue();
        }
    }

    xhr.send(form);
}

function prepareForm(json, file) {
    var form = new FormData();
    
    form.append("key", json["key"] + "/" + file.name);
    form.append("X-Amz-Credential", json["x-amz-credential"]);
    form.append("X-Amz-Algorithm", "AWS4-HMAC-SHA256");
    form.append("X-Amz-Date", json["x-amz-date"]);
    form.append("policy", json["policy"]);
    form.append("X-Amz-Signature", json["signature"]);
    form.append("file", file, file.name);

    return form;
}

function uploadItem(file, textSpan, progressSpan, progressBar, json) {
    var data = prepareForm(json, file);

    var xhr = new XMLHttpRequest();
    xhr.open("POST", "http://" + json["bucket"] + ".s3.amazonaws.com/");
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
        if (xhr.status == 204) {
            progressBar.setAttribute("class", "progress-bg progress-complete-bg");

            var linkNode = document.createElement("A");
            linkNode.setAttribute("href", "/view/" + json["key"] + "/" + json["upload_id"]);
            linkNode.setAttribute("target", "_blank");
            linkNode.appendChild(document.createTextNode(file.name));
            clearChildren(textSpan);
            textSpan.appendChild(linkNode);

            progressSpan.innerHTML = "Complete";

            finishUpload(json.upload_id)
        } else {
            console.log(xhr.responseText);
            progressBar.setAttribute("class", "progress-bg progress-failure-bg");
            progressSpan.innerHTML = "Failed";
        }
        queue.splice(0, 1);
        uploadFromQueue();
    }

    xhr.send(data);
}

function finishUpload(id) {
    form = new FormData();
    form.append("id", id);
    xhr = new XMLHttpRequest();
    xhr.open("POST", "/upload_confirm");
    xhr.onload = function() {
        if (xhr.status != 200) {
            alert("Warning: unable to confirm upload");
        }
    }

    xhr.send(form);
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
    for (var i = 0; i < files.length; i++) {
        uploadFile(files[i]);
    }

    if (!uploadInProgress)
        uploadFromQueue();
}
