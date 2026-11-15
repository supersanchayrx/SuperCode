SuperCode - a small AI coding agent for the terminal
====================================================

INSTALL
  1. Unzip this folder anywhere.
  2. Double-click install.cmd
  3. Open a NEW terminal (PowerShell / cmd / VS Code terminal).

FIRST RUN
  Go to a project folder and run:   supercode
  It will ask for an OpenRouter API key (free at https://openrouter.ai/keys)
  and save it to %USERPROFILE%\.supercode\.env
  You can edit that file later to add more keys (SUPERCODE_API_KEY2, ...)
  or pick different models.

USING IT
  supercode                 start (auto mode picks chat / plan / edit for you)
  supercode -M chat         start in a specific mode: chat, plan or edit
  supercode --version

  Inside SuperCode:
    /help      all commands
    /plan ...  write a plan into PLAN.md
    /edit ...  let it change files (it shows a diff and asks y/N first)
    /status    models and key status
    /exit

  SuperCode only reads and edits files inside the folder you start it in.
  Every file change and command needs your y/N approval.

NOT INSTALLING?
  You can also just run supercode.exe directly from this folder.

UNINSTALL
  Double-click uninstall.cmd

NOTE
  Windows may show a SmartScreen warning because the exe is not signed.
  Click "More info" -> "Run anyway".
