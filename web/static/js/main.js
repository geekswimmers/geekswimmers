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

                    createSelectOption(eventsField, "0", "Select...", false);

                    events.forEach(event => {
                        createSelectOption(eventsField,
                            event.ID,
                            event.Distance + 'm ' + toTitleCase(event.Style.Stroke),
                            false
                        );
                    });
                })
                .catch(error => {
                    console.error('Error:', error);
                });
            break;
        }
    }
}

function updateTeamsByJurisdiction(selectedId) {
    const jurisdictionCbx = document.getElementById("jurisdiction");
    const teamCbx = document.getElementById("team");
    teamCbx.innerHTML = "";
    fetch('/api/teams/?jurisdiction=' + jurisdictionCbx.value)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(teams => {
            createSelectOption(teamCbx, "0", "Select...", false);

            if (teams != null) {
                teams.forEach(team => {
                    const selected = team.ID.Int64 === selectedId;
                    createSelectOption(teamCbx, team.ID.Int64, team.Acronym +" - "+ team.FullName, selected);
                });
            }
        });
}

function createSelectOption(selectField, value, text, selected) {
    const option = document.createElement("option");
    option.value = value;
    option.text = text;
    option.selected = selected;
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

function deleteSwimmer(swimmerId) {
    fetch('/profile/swimmers/'+ swimmerId +'/', { method: 'DELETE' })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok' + response.Error);
            }
            location.href = '/profile/';
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

function formatMilliseconds(milliseconds) {
    let ms = milliseconds;
    let signal = "";

    if (milliseconds < 0) {
        ms = -milliseconds;
        signal = "-";
    }

    // Logic equivalent to FromMilliseconds
    const minutes = Math.floor(ms / 60000);
    const seconds = Math.floor((ms % 60000) / 1000);
    // Calculate centiseconds (milliseconds / 10)
    const centiseconds = Math.floor((ms % 1000) / 10);

    // Logic equivalent to FormatTime (using padStart for %02d)
    const minStr = minutes.toString().padStart(2, '0');
    const secStr = seconds.toString().padStart(2, '0');
    const csStr = centiseconds.toString().padStart(2, '0');

    return `${signal}${minStr}:${secStr}.${csStr}`;
}