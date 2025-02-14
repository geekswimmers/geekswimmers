const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
const tooltipList = [...tooltipTriggerList].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))

function updateEventsByCourse(swimmerId) {
    const course = document.getElementsByName('course');
    for (let i = 0; i < course.length; i++) {
        if (course[i].checked) {
            fetch('/api/swimmers/'+ swimmerId +'/events/?course=' + course[i].value)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(events => {
                    const eventsField = document.getElementById("event");
                    eventsField.innerHTML = "";

                    createSelectOption(eventsField, "0", "Select...");

                    events.forEach(event => {
                        createSelectOption(eventsField,
                            event.ID,
                            event.Distance + 'm ' + toTitleCase(event.Style.Stroke));
                    });
                })
                .catch(error => {
                    console.error('Error:', error);
                });
            break;
        }
    }
}

function createSelectOption(selectField, value, text) {
    const option = document.createElement("option");
    option.value = value;
    option.text = text;
    selectField.appendChild(option);
}

function toTitleCase(str) {
    return str.replace(
        /\w\S*/g,
        text => text.charAt(0).toUpperCase() + text.substring(1).toLowerCase()
    );
}

function deleteBestTime(swimmerId, bestTimeId) {
    fetch('/profile/swimmers/'+ swimmerId +'/besttimes/'+ bestTimeId +'/', { method: 'DELETE' })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            location.href = '/profile/swimmers/'+ swimmerId +'/';
        })
        .catch(error => {
            console.error('Error:', error);
        });
}

function acceptLinkRequest(swimmerId, linkId) {
    fetch('/api/profile/swimmers/'+ swimmerId +'/parentlink/'+ linkId +'/', { method: 'PUT' })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

function dismissLinkRequest(swimmerId, linkId) {
    fetch('/api/profile/swimmers/'+ swimmerId +'/parentlink/'+ linkId +'/', { method: 'DELETE' })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}