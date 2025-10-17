# RPS (Robot Pits Helper)
Manage your FRC robot code the UNIX way.

## Installing
You can grab the latest release [here](https://github.com/Squibid/rph/releases/latest).
Don't worry about the version, all versions should be compatible with all
versions of wpilib.

## Quick Start
Make a new project from a template
```sh
rph template -d MyNewRoboProject -l java -t commandbased -n 5438 -s false
```
If you'd like to learn more about the template subcmd just run `rph template -h`.
Now let's go into the project and add a vendor dependency:
```sh
rph vendordep add photonlib-2025.3.1
```
And now we just need to build to make it download all the actual code:
```sh
gradle build
```

## But Why?
Well you probably shouldn't, but if you really want to learn more about how
computers work this is a solid starting point. As for the creation of this
project: it was made for two reasons:
1. To alleviate my everlasting frustration with WPILIB's crappy uis
2. To dust everyone in the YAMS speedrunning competition

## TODO
- [x] template
- [ ] vendordep
    - [ ] update installed vendor deps
    - [ ] get info about installed vendor deps
- [ ] riolog listener, seems like the vscode extension does it which means
      there's no reason we can't >:)

- [ ] have to make sure wpilib is installed somewhere so that we can build
      projects without running the wpilib installer, this might take a bit
      of investigation
- [ ] make a declaritive config? (really only seems useful for speedrunning or
      setting up multiple rookies machines for a new project)

- [ ] runner? (maybe idk though cause you can just use gradlew directly)
      no, but maybe we should document how to use gradle for the noobs who
      use the internet
