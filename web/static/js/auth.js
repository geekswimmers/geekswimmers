function checkStrength(password) {
    let mediumRegex = new RegExp("^(((?=.*[a-z])(?=.*[A-Z]))|((?=.*[a-z])(?=.*[0-9]))|((?=.*[A-Z])(?=.*[0-9])))(?=.{6,})");
    let strongRegex = new RegExp("^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%^&\*])(?=.{8,})");

    let passwordStrength = document.getElementById("password-strength");
    let passwordStrengthLabel = document.getElementById("password-strength-label");

    if(strongRegex.test(password)) {
        passwordStrength.src = "/static/images/password-strong.svg";
        passwordStrengthLabel.textContent = "Strong";
    } else if(mediumRegex.test(password)) {
        passwordStrength.src = "/static/images/password-regular.svg";
        passwordStrengthLabel.textContent = "Regular";
    } else if (password.length === 0) {
        passwordStrength.src = "/static/images/password-empty.svg";
        passwordStrengthLabel.textContent = "Empty";
    } else {
        passwordStrength.src = "/static/images/password-weak.svg";
        passwordStrengthLabel.textContent = "Weak";
    }
}

let passwordField = document.getElementById("password");
if (passwordField) {
    passwordField.addEventListener("keyup", function () {
        let password = document.getElementById("password").value;
        let confirmPassword = document.getElementById("confirm-password").value;
        let submitBtn = document.getElementById("submitBtn");
    
        if (password === confirmPassword) {
            submitBtn.disabled = false;
        } else {
            submitBtn.disabled = true;
        }
    
        checkStrength(password);
    });
    checkStrength(passwordField.value);
}

let confirmPasswordField = document.getElementById("confirm-password");
if (confirmPasswordField) {
    confirmPasswordField.addEventListener("keyup", function () {
        let password = document.getElementById("password").value;
        let confirmPassword = document.getElementById("confirm-password");
        let submitBtn = document.getElementById("submitBtn");
    
        if (password === confirmPassword.value) {
            confirmPassword.classList.add("is-valid");
            submitBtn.disabled = false;
        } else {
            confirmPassword.classList.remove("is-valid");
            submitBtn.disabled = true;
        }
    });
}

let btnSignoff = document.getElementById("btn-signoff");
if (btnSignoff) {
    btnSignoff.addEventListener("click", function () {
        if (confirm("Are you sure you want to sign-off?")) {
            return true;
        }
        return false;
    });
}

let chkPromotionalNotification = document.getElementById("promotional-notification");
if (chkPromotionalNotification) {
    chkPromotionalNotification.addEventListener("click", function () {
        let xhr = new XMLHttpRequest();
        xhr.open("PUT", "", true);
        xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4 && xhr.status === 200) {
                let successMsg = document.getElementById("notifications-saved");
                successMsg.innerHTML = "&nbsp;Choices saved!";
                setTimeout(function () {
                    successMsg.innerHTML = "&nbsp;";
                }, 5000);
            }
        };
        xhr.send("notification_promo=" + chkPromotionalNotification.checked);
    });
}