package org.baf.gol;

import java.util.ArrayList;
import java.util.Arrays;

/**
 * Main GoL engine.
 */
public class Main implements ParameterProvider {
  // command line values
  String urlString, nameString;
  int magFactorInt = 1, gameCycles = 10;
  boolean startServerFlag, runTimingsFlag, reportFlag, saveImageFlag;
  public static String saveImageRoot = "/temp"; // change per OS type

  @Override
  public String getUrlParameter() {
    return urlString;
  }

  @Override
  public String getNameParameter() {
    return nameString;
  }

  @Override
  public int getMagFactorParameter() {
    return magFactorInt;
  }

  @Override
  public boolean startServerFlag() {
    return startServerFlag;
  }

  @Override
  public boolean runTimingsFlag() {
    return runTimingsFlag;
  }

  @Override
  public boolean reportFlag() {
    return reportFlag;
  }

  @Override
  public boolean saveImageFlag() {
    return saveImageFlag;
  }

  @Override
  public int getGameCyclesParameter() {
    return gameCycles;
  }

  /**
   * Main entry point.
   * 
   * Sample: -n tiny1 -u file:/.../tiny1.png
   */
  public static void main(String[] args) {
    if (args.length == 0) {
      printHelp();
      return;
    }
    try {
      var main = new Main();
      if (!main.parseArgs(args)) {
        Logger.log.tracef("Command arguments: %s", Arrays.toString(args));
        printHelp();
        System.exit(1);
      }
      main.launch();
    } catch (Exception e) {
      Logger.log.exceptionf(e, "launched failed");
      System.exit(3);
    }
  }

  private void launch() throws Exception {
    Game.coreGame = new Game(this);
    Game.coreGame.saveImageRoot = saveImageRoot;
    Game.coreGame.maxCycles = gameCycles;

    // need timings
    if (!urlString.isEmpty()) {
      if (nameString.isEmpty()) {
        System.err.printf("a name is required when a URL is provided%n");
        System.exit(1);
      }
      if (runTimingsFlag) {
        runCycleTimings();
      }
    }

    // need server
    if (startServerFlag) {
      // launch HTTP server
      var server = new Server(this);
      server.saveImageRoot = saveImageRoot;
      server.startHttpServer();
    }
  }

  // approximation of flag package in Go
  private boolean parseArgs(String[] args) {
    boolean ok = true;
    try {
      for (int i = 0; i < args.length; i++) {
        switch (args[i].toLowerCase()) {
          case "-url":
          case "-u":
            urlString = args[++i];
            break;
          case "-name":
          case "-n":
            nameString = args[++i];
            break;
          case "-magfactor":
          case "-mf":
          case "-mag":
            magFactorInt = Integer.parseInt(args[++i]);
            if (magFactorInt < 1 || magFactorInt > 20) {
              throw new IllegalArgumentException("bad magFactor: " + magFactorInt);
            }
            break;
          case "-gamecycles":
          case "-gc":
            gameCycles = Integer.parseInt(args[++i]);
            if (gameCycles < 1 || gameCycles > 1000) {
              throw new IllegalArgumentException("bad gameCycles: " + gameCycles);
            }
            break;
          case "-start":
            startServerFlag = true;
            break;
          case "-time":
            runTimingsFlag = true;
            break;
          case "-report":
            reportFlag = true;
            break;
          case "-saveimage":
          case "-si":
            saveImageFlag = true;
            break;
          default:
            throw new IllegalArgumentException("unknown parameter key: " + args[i]);
        }
      }
    } catch (Exception e) {
      System.err.printf("parse failed: %s%n", e.getMessage());
      ok = false;
    }
    return ok;
  }

  // get execution timings
  private void runCycleTimings() throws Exception {
    var cpuCount = Runtime.getRuntime().availableProcessors();
    for (var i = 1; i <= 64; i *= 2) {
      Logger.log.tracef("Running with %d threads, %d CPUs...", i, cpuCount);
      Game coreGame = Game.coreGame;
      coreGame.threadCount = i;
      coreGame.run(getNameParameter(), getUrlParameter());

      if (reportFlag()) {
        Logger.log.tracef("Game max: %d, go count: %d:", i, coreGame.maxCycles, coreGame.threadCount);
        for (var grk : coreGame.runs.keySet()) {
          var gr = coreGame.runs.get(grk);
          Logger.log.tracef("Game Run: %s, cycle count: %d", gr.name, gr.cycles.size());
          for (var c : gr.cycles) {
            long start = c.startedAt.getTime(), end = c.endedAt.getTime();
            Logger.log.tracef("Cycle: start epoch: %dms, end epoch: %dms, elapsed: %dms", start, end, end - start);
          }
        }
      }
    }
  }

  private static void printHelp() {
    System.err.printf("%s%n%n%s%n", trimWhitespace(golDescription), trimWhitespace((golArgs)));
  }

  private static Object trimWhitespace(String lines) {
    var xlines = lines.split("\n");
    var result = new ArrayList<String>();
    for (int i = 0, c = xlines.length; i < c; i++) {
      String tline = xlines[i].trim();
      if (!tline.isEmpty()) {
        result.add(tline.replace("%n", "\n"));
      }
    }
    return String.join("\n", result);
  }

  static String golDescription = """
       Play the game of Life.
       Game boards are initialized from PNG images.
       Games play over several cycles.%n
       Optionally acts as a server to retrieve images of game boards during play.%n
       No supported positional arguments.
      """;

  static String golArgs = """
      Arguments (all start with '-'):
      url|u <url>              URL of the PNG image to load
      name|n <name>            name to refer to the game initialized by the URL
      magFactor|mf|mag <int>   magnify the grid by this factor when formatted into an image  (default 1; 1 to 20)
      gameCycles|gc <int>      sets number of cycles to run (default 10)
      start <boolean>          start the HTTP server (default false)
      time <boolean>           run game cycle timings with different thread counts (default false)
      report <boolean>         output run statistics (default false)
      saveImage|si <boolean>   save generated images into a file (default false)
      """;
}
