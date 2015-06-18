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
