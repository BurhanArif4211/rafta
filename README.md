# Rafta
# Rafta

Rafta is a note taking and todo list application that runs entirely on your local machine and syncs across devices over your home network. 

This is an app for myself since i am tired of samsung notes not properly syncying with onenote. And I also wanted to bring the functionality of MS TODO and samsung notes to one app. 

* **WARNING**: As an information security enthusiast, i urge you to not use this in a production enviroment. The app exposes your device open on the local network without any authentication or encryption. Use at your own risk. I consider my local network secure that's why i haven't added security features yet.

# Showcase
<div style="display: flex; gap:4px; justify-content: center;"> 
<img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/2.jpg" alt="App screenshot 2" style=" max-width: 300px; width: 100%; height: auto;" /> 
<img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/4.jpg" alt="App screenshot 4" style=" max-width: 300px; width: 100%; height: auto;" /> </div>


## How it works

The notes editor lets you write in Markdown and switch to a preview. 
For todos, you can break down tasks into steps, check them off as you go.

All your data is stored locally in a **SQLite database**. 

When you want to sync between devices, you open the sync dialog, enter the IP address of another device running Rafta on your network, and pull the latest data. The app replaces its own database with what it gets from the other device, so you always have the version you intentionally chose. It is not automatic, **and there is no conflict resolution – you decide when and what to sync.**

## Notes
* I don't like to grind code for hours and software dev is not my focus in career, so most of the implementation is done by deepseek. It is probably the best model out there to write code. The architecture is still mainly planned by me. 
* The whole point of this is to save time by not having to open multiple apps and instant syncing between my PC, chromebook (linux) and phone in the LAN

## Technologies used

Rafta is written in Go and uses the Fyne toolkit for its graphical interface. Data is stored in `SQLite`. The sync part runs a tiny HTTP server on `4211` port and the client fetches everything as `JSON`.

- Fyne for the cross‑platform UI
- SQLite for local storage 

## Building from source

If you want to build Rafta yourself, you need Go 1.20 or newer. Clone the repository and run:
- First you need to make sure you have fyne apps compilable on you current enviroment. See: (DOCS)[https://docs.fyne.io/started/quick/]

```batch
git clone https://github.com/BurhanArif4211/rafta
cd rafta
.\script\build.bat 
```
Above script builds for windows and Android (ran from windows).
## Contributing

Rafta is a personal project but I am open to contributions. If you find a bug, have an idea for a feature, or want to improve the documentation, feel free to open an issue or a pull request on the repository.

Before submitting anything, please keep in mind that the goal is to keep the app simple and **local‑first**. Features that depend on external services, but discussions are always welcome.

*Be Realistic*