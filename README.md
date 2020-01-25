# RMKDump

A Remarkable tablet backup utility written in Go

This is a very simple (but efficient) backup too for remarkable.
Basically, simply run "./rmkdump <optional local backup directory>" and all documents from your Remarkable will be exported as PDF to the target folder (default is named "./backup/").
The tool wil detect unchanged files since last backup and ignore them.

# Building

    go build

# How does it work

This tool uses the USB web interface, thus make sure your remarkable is connected to your computer and the USB Web interface is turned on.
"rmkdump" connects to the web APIs and crawls the complete document hierarchy.
An index containing the document ID and last modified version is saved at the backup directory root in order to optimize sync.

# Support

Not much :)
