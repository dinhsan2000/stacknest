<?php
// Stacknest Adminer wrapper — allows passwordless local login

function adminer_object() {
    class StacknestAdminer extends Adminer {
        function login($login, $password) {
            return true; // allow empty password for local dev
        }
    }
    return new StacknestAdminer;
}

require __DIR__ . '/adminer.php';
