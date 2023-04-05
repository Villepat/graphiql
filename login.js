$(document).ready(function() {
    $('#login-form').submit(function(e) {
        e.preventDefault();
        var identifier = $('#identifier').val();
        var password = $('#password').val();
        console.log('button piushed')
        $.ajax({
            url: '/login',
            type: 'POST',
            dataType: 'json',
            data: {
                identifier: identifier,
                password: password
            },
            success: function(response) {
                // Handle successful login
            },
            error: function(jqXHR, textStatus, errorThrown) {
                // Handle login error
            }
        });
    });
});
