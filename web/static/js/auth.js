function checkStrength(password) {
    let mediumRegex = new RegExp("^(((?=.*[a-z])(?=.*[A-Z]))|((?=.*[a-z])(?=.*[0-9]))|((?=.*[A-Z])(?=.*[0-9])))(?=.{6,})");
    let strongRegex = new RegExp("^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#\$%\^&\*])(?=.{8,})");

    if(strongRegex.test(password)) {
        $("#password-strength").attr("src", "/static/images/strong.svg");
        $("#password-strength-label").text("Strong");
    } else if(mediumRegex.test(password)) {
        $("#password-strength").attr("src", "/static/images/regular.svg");
        $("#password-strength-label").text("Regular");
    } else {
        $("#password-strength").attr("src", "/static/images/weak.svg");
        $("#password-strength-label").text("Weak");
    }
}

$("#password").keyup(function () {
    let password = $("#password").val();
    let confirmPassword = $("#confirm-password");
    let submitBtn = $("#submitBtn");

    if (password === confirmPassword.val()) {
        submitBtn.prop('disabled', false);
    } else {
        submitBtn.prop('disabled', true);
    }

    checkStrength(password);
});

$("#confirm-password").keyup(function () {
    let password = $("#password");
    let confirmPassword = $("#confirm-password");
    let submitBtn = $("#submitBtn");

    if (password.val() === confirmPassword.val()) {
        confirmPassword.addClass("is-valid");
        submitBtn.prop('disabled', false);
    } else {
        confirmPassword.removeClass("is-valid");
        submitBtn.prop('disabled', true);
    }
});

$("#btn-signoff").click(function () {
    if(confirm("Are you sure you want to sign-off?")) {
        return true;
    }
    return false;
});

$("#promotional-notification").click(function() {
    $.ajax({
        url: "",
        type: "PUT",
        data: "notification_promo=" + $(this).prop('checked'),
        success: function () {
            let successMsg = $("#notifications-saved");
            successMsg.html("&nbsp;Choices saved!");
            setTimeout( function () {
                successMsg.html("&nbsp;");
            }, 5000);
        }
    });
});