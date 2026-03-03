# Rafta

Rafta is a note taking and todo list application that runs entirely on your local machine and syncs across devices over your home network. 

This is an app for myself since i am tired of samsung notes not properly syncying with onenote. And I also wanted to bring the functionality of MS TODO and samsung notes to one app. 

# Showcase

<div style="display: flex; justify-content: space-between; align-items: flex-start;">
  <!-- <img src="url_to_your_image1" alt="Image 1" style="width: 60%; max-width: 600px;"/> -->
  <div style="display: flex; flex-direction: column; justify-content: space-between;">
    <img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/1.png" alt="Image 2" style="width: 40%; max-width: 200px; margin-bottom: 5px;"/>
    <img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/2.jpg" alt="Image 3" style="width: 40%; max-height: 200px;"/>
  </div>
</div>

<div style="display: flex; justify-content: space-between; align-items: flex-start; margin-top: 20px;">
  <img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/3.png" alt="Image 4" style="width: 60%; max-width: 600px;"/>
  <img src="https://raw.githubusercontent.com/BurhanArif4211/rafta/refs/heads/main/img/4.jpg" alt="Image 5" style="width: 40%; max-height: 200px;"/>
</div>


## How it works

The notes editor lets you write in Markdown and switch to a preview. 
For todos, you can break down tasks into steps, check them off as you go.

All your data is stored locally in a **SQLite database**. 

When you want to sync between devices, you open the sync dialog, enter the IP address of another device running Rafta on your network, and pull the latest data. The app replaces its own database with what it gets from the other device, so you always have the version you intentionally chose. It is not automatic, and there is no conflict resolution – you decide when and what to sync.

## Notes
* I don't like to grind code for hours and software dev is not my focus in career, so most of the implementation is done by deepseek. It is probably the best model out there to write code. The architecture is still mainly planned by me.
* I plan to make an android app for this too. 
* The whole point of this is to save time by not having to open multiple apps and instant syncing between my pc, chromebook (linux) and phone in the LAN

## Technologies used

Rafta is written in Go and uses the Fyne toolkit for its graphical interface. Data is stored in SQLite with UUIDs as primary keys to keep things unique across devices. The sync part runs a tiny HTTP server on a configurable port and the client fetches everything as JSON.

- Go for the core logic and concurrency
- Fyne for the cross‑platform UI
- SQLite for local storage 

## Building from source

If you want to build Rafta yourself, you need Go 1.20 or newer. Clone the repository and run:

```
go build -o rafta ./cmd/rafta
```

That will produce an executable for your platform. You can also cross‑compile for other operating systems by setting the GOOS and GOARCH environment variables.

## Contributing

Rafta is a personal project but I am open to contributions. If you find a bug, have an idea for a feature, or want to improve the documentation, feel free to open an issue or a pull request on the repository.

Before submitting anything, please keep in mind that the goal is to keep the app simple and **local‑first**. Features that depend on external services or complicate the sync model are unlikely to be merged, but discussions are always welcome.
*Be Realistic*
