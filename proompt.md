I am making a note taking + todo app that runs fully locally and syncs my devices over the LAN (since i already have static ips on my all devices, i can add configuration in the app to set ips). Lets plan out the project properly before implementing. I want it to be a modular design which is easily manageable and extendable for future.   

# Basic Requirements
## Note Taking
- The inspiration for note taking is samsung notes mobile app. It has a folder structure in which we can add notes and manage documents in an organized manner.
- note taker will be markdown editor which will allow bullets, bolding/italicizing/font sizes,other coloring features and more.

## TODO Taking
- ToDo section will have namable folders and each section can have todos. each todo will be able to have steps. we currently don't need time and scheduling stuff later maybe.

## Syncing
- I plan to have a config section where i can add devices which are syncing in the local network. 
- There are soo many edge cases when it comes to syncing. 
- We could expose a specific port on the local network, and let other devices listen for connections and give options on all devices to sync with a specific device (I don't work on same documents or todos at the same time). 
- For instance, if a pc has some data, i go out and take some notes on my laptop, make todos, when i get home, i should be able to open my pc and sync the todos and notes to my pc. It is not like version control, i decide whenever i want fetch sync from a device and make that version of the database canonical.
- 
