const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
const tooltipList = [...tooltipTriggerList].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))

//if (location.protocol != "https:" && location.hostname != "localhost") {
//    window.location.replace("https://www.geekswimmers.com");
//}
