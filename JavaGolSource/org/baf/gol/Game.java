package org.baf.gol;

import java.awt.GridLayout;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;
import java.io.BufferedOutputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Date;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import javax.imageio.ImageIO;
import javax.imageio.stream.MemoryCacheImageOutputStream;
import javax.swing.ImageIcon;
import javax.swing.JFrame;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.border.TitledBorder;

import org.baf.gol.Server.ResponseDataSender;

/**
 * Represents a GoL Game with a set of Game runs.
 */
public class Game {
  public static Game coreGame; // global instance

  static int threadId;

  private int nextThreadId() {
    return ++threadId;
  }

  // play history
  public Map<String, GameRun> runs = new LinkedHashMap<>();

  public int maxCycles = 25; // max that can be played
  public int threadCount; // thread to use in timings
  ParameterProvider pp; // source of command line parameters
  public String saveImageRoot = "/temp"; // change per OS type

  public Game(ParameterProvider pp) {
    this.pp = pp;
  }

  /**
   * Represents a single run of a GoL Game.
   */
  public class GameRun {
    static final int offIndex = 255, onIndex = 0;
    static final int midvalue = 256 / 2; // separates black vs. white

    public Game parent;
    public String name;
    public String imageUrl;
    public Date startedAt, endedAt;
    public int width, height;
    public Grid initialGrid, currentGrid, finalGrid;
    public List<GameCycle> cycles = new ArrayList<>();
    public int delayIn10ms, playIndex;
    public int threadCount;

    private String author = "Unknown";

    public String getAuthor() {
      return author;
    }

    public void setAuthor(String author) {
      this.author = author;
    }

    public GameRun(Game game, String name, String url) throws Exception {
      this.parent = game;
      this.name = name;
      this.imageUrl = url;
      this.delayIn10ms = 5 * 100;

      // make the game grid and load initial state
      String[] kind = new String[1];
      BufferedImage img = Utility.loadImage(url, kind);
      Logger.log.tracef("Image kind: %s", kind[0]);
      if (!"png".equals(kind[0].toLowerCase())) {
        throw new IllegalArgumentException(String.format("named image %s is not a PNG", url));
      }
      var bounds = new Rectangle(img.getMinX(), img.getMinY(), img.getWidth(), img.getHeight());
      var size = bounds.getSize();
      initialGrid = new Grid(size.width, size.height);
      width = initialGrid.width;
      height = initialGrid.height;
      initGridFromImage(bounds.x, bounds.y, bounds.width, bounds.height, img);
      currentGrid = initialGrid.deepClone();
    }

    @Override
    public String toString() {
      return "GameRun[name=" + name + ", imageUrl=" + imageUrl + ", startedSt=" + startedAt + ", endedAt=" + endedAt
          + ", width=" + width + ", height=" + height + ", cycles=" + cycles + ", delayIn10ms=" + delayIn10ms
          + ", playIndex=" + playIndex + ", threadCount=" + threadCount + "]";
    }

    private void initGridFromImage(int minX, int minY, int maxX, int maxY, BufferedImage img) {
      for (int y = minY; y < maxY; y++) {
        for (int x = minX; x < maxX; x++) {
          var pixel = img.getRGB(x, y);
          int r = (pixel >> 16) & 0xFF, g = (pixel >> 8) & 0xFF, b = (pixel >> 0) & 0xFF;

          var cv = 0; // assume all dead
          if (r + g + b < midvalue * 3) {
            cv = 1; // make cell alive
          }
          initialGrid.setCell(x, y, cv);
        }
      }
    }

    public void sendPng(ResponseDataSender rs, int index, int mag) throws IOException {
      Grid grid = null;
      switch (index) {
        case 0:
          grid = initialGrid;
          break;
        default:
          index--;
          if (index < 0 || index >= cycles.size()) {
            throw new ArrayIndexOutOfBoundsException("bad index");
          }
          grid = cycles.get(index).afterGrid;
      }

      var img = new BufferedImage(width * mag + 1, height * mag + 1, BufferedImage.TYPE_BYTE_BINARY);
      fillImage(grid, mag, img);
      var b = encodePngImage(img);
      rs.sendResponseData(b);
      showImageInGui(img); // show in GUI

      if (parent.pp.saveImageFlag()) {
        var saveFile = String.format(saveImageRoot + "/Image_%s.gif", name);
        Files.write(Paths.get(saveFile), b);
        Logger.log.tracef("Save %s", saveFile);
      }
    }

    private byte[] encodePngImage(BufferedImage img) throws IOException {
      var baos = new ByteArrayOutputStream();
      var bos = new BufferedOutputStream(baos);
      var ios = new MemoryCacheImageOutputStream(bos);
      ImageIO.write(img, "png", ios);
      ios.flush();
      return baos.toByteArray();
    }

    private void fillImage(Grid grid, int mag, BufferedImage img) {
      for (var row = 0; row < grid.height; row++) {
        for (var col = 0; col < grid.width; col++) {
          var index = grid.getCell(col, row) != 0 ? onIndex : offIndex;
          // apply magnification
          for (var i = 0; i < mag; i++) {
            for (var j = 0; j < mag; j++) {
              img.setRGB(mag * col + i, mag * row + j, index == onIndex ? 0 : 0x00FFFFFF);
            }
          }
        }
      }
    }

    /**
     * Run a game.
     */
    public void run() {
      this.threadCount = coreGame.threadCount;
      startedAt = new Date();
      int maxCycles = parent.maxCycles;
      for (int count = 0; count < maxCycles; count++) {
        nextCycle();
      }
      endedAt = new Date();
      Logger.log.tracef("GameRun total time: %dms, cycles: %d, thread count: %d",
          endedAt.getTime() - startedAt.getTime(), maxCycles, threadCount);
      finalGrid = currentGrid.deepClone();
    }

    // Advance and play next game cycle.
    // Updating of cycle grid rows can be done in parallel;
    // which can reduce execution time.
    private void nextCycle() {
      var gc = new GameCycle(this);
      gc.beforeGrid = currentGrid.deepClone();
      var p = gc.parent;
      var threadCount = Math.max(p.parent.threadCount, 1);
      gc.afterGrid = new Grid(gc.beforeGrid.width, gc.beforeGrid.height);
      gc.startedAt = new Date();
      var threads = new ArrayList<Thread>();
      var rowCount = (height + threadCount / 2) / threadCount;
      for (var i = 0; i < threadCount; i++) {
        var xi = i;
        var t = new Thread(() -> {
          procesRows(gc, rowCount, xi * rowCount, gc.beforeGrid, gc.afterGrid);
        }, "thread_" + nextThreadId());
        threads.add(t);
        t.setDaemon(true);
        t.start();
      }
      for (var t : threads) {
        try {
          t.join();
        } catch (InterruptedException e) {
          // ignore
        }
      }
      gc.endedAt = new Date();
      currentGrid = gc.afterGrid.deepClone();
      cycles.add(gc);
      gc.cycleCount = cycles.size();
    }

    // process all cells in a set of rows
    private void procesRows(GameCycle gc, int rowCount, int startRow, Grid inGrid, Grid outGrid) {
      for (var index = 0; index < rowCount; index++) {
        var rowIndex = index + startRow;
        for (var colIndex = 0; colIndex < width; colIndex++) {
          // count any neighbors
          var neighbors = 0;
          if (inGrid.getCell(colIndex - 1, rowIndex - 1) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex, rowIndex - 1) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex + 1, rowIndex - 1) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex - 1, rowIndex) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex + 1, rowIndex) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex - 1, rowIndex + 1) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex, rowIndex + 1) != 0) {
            neighbors++;
          }
          if (inGrid.getCell(colIndex + 1, rowIndex + 1) != 0) {
            neighbors++;
          }
          // determine next generation cell state based on neighbor count
          var pv = inGrid.getCell(colIndex, rowIndex);
          var nv = 0;
          switch (neighbors) {
            case 2:
              nv = pv;
              break;
            case 3:
              if (pv == 0) {
                nv = 1;
              }
              break;
          }
          outGrid.setCell(colIndex, rowIndex, nv);
        }
      }
    }

    /**
     * Make images from 1+ cycles into GIF form.
     */
    public byte[] makeGifs(int count, int mag) throws IOException {
      var cycleCount = cycles.size();
      var xcycles = Math.min(count, cycleCount + 1);
      List<BufferedImage> bia = new ArrayList<>();
      var added = addGridSafe(initialGrid, 0, xcycles, mag, bia);
      for (int i = 0; i < cycleCount; i++) {
        added = addGridSafe(cycles.get(i).afterGrid, added, xcycles, mag, bia);
      }
      return packGifs(added, mag, delayIn10ms, bia.toArray(new BufferedImage[bia.size()]));
    }

    int addGridSafe(Grid grid, int added, int max, int mag, List<BufferedImage> bia) {
      var img = new BufferedImage(mag * width + 1, mag * height + 1, BufferedImage.TYPE_BYTE_BINARY);
      if (added < max) {
        fillImage(grid, mag, img);
        bia.add(img);
        added++;
      }
      return added;
    }

    byte[] packGifs(int count, int mag, int delay, BufferedImage[] bia) throws IOException {
      showImagesInGui(bia);

      var baos = new ByteArrayOutputStream();
      var bos = new BufferedOutputStream(baos);
      var ios = new MemoryCacheImageOutputStream(bos);
      AnnimatedGifWriter.createGifs(ios, delay, author, bia);
      ios.flush();
      return baos.toByteArray();
    }

    // not in Go version.
    void showImagesInGui(BufferedImage[] bia) {
      // create a Swing Frame to show a row of images
      var frame = new JFrame("Show Images rendered at " + new Date());
      frame.setDefaultCloseOperation(JFrame.DISPOSE_ON_CLOSE);
      JPanel imagePanel = new JPanel(new GridLayout());
      var sp = new JScrollPane(imagePanel);
      frame.setContentPane(sp);
      frame.setSize(1000, 800);

      var index = 1;
      for (var bi : bia) {
        var icon = new ImageIcon(bi);
        JLabel labelledIcon = new JLabel(icon);
        labelledIcon.setBorder(
            new TitledBorder(String.format("Image: %d (%dx%d)", index++, icon.getIconWidth(), icon.getIconHeight())));
        imagePanel.add(labelledIcon);
      }
      frame.setVisible(true);
    }

    // not in Go version.
    void showImageInGui(BufferedImage bi) {
      var frame = new JFrame("Show Image rendered at " + new Date());
      JPanel imagePanel = new JPanel(new GridLayout());
      var sp = new JScrollPane(imagePanel);
      frame.setContentPane(sp);
      frame.setDefaultCloseOperation(JFrame.DISPOSE_ON_CLOSE);
      frame.setSize(1000, 800);
      var icon = new ImageIcon(bi);
      JLabel labelledIcon = new JLabel(icon);
      labelledIcon
          .setBorder(new TitledBorder(String.format("Image: (%dx%d)", icon.getIconWidth(), icon.getIconHeight())));
      imagePanel.add(labelledIcon);
      frame.setVisible(true);
    }
  }

  /**
   * Clear all runs.
   */
  public void clear() {
    runs.clear();
  }

  /**
   * Run a game.
   */
  public void run(String name, String url) throws Exception {
    var gr = new GameRun(this, name, url);
    runs.put(gr.name, gr);
    gr.run();
  }

  /**
   * Represents a GoL Game grid.
   */
  public static class Grid {
    public byte[] data;
    public int width, height;

    public Grid(int width, int height) {
      this.width = width;
      this.height = height;
      data = new byte[width * height];
    }

    @Override
    public String toString() {
      return "Grid[width=" + width + ", height=" + height + "]";
    }

    public int getCell(int x, int y) {
      if (x < 0 || x >= width || y < 0 || y >= height) {
        return 0;
      }
      return data[x + y * width];
    }

    public void setCell(int x, int y, int cv) {
      if (x < 0 || x >= width || y < 0 || y >= height) {
        return;
      }
      data[x + y * width] = (byte) cv;
    }

    public Grid deepClone() {
      var ng = new Grid(width, height);
      for (int i = 0; i < data.length; i++) {
        ng.data[i] = data[i];
      }
      ng.width = width;
      ng.height = height;
      return ng;
    }
  }

  /**
   * Represents a GoL Game cycle.
   */
  public static class GameCycle {
    public GameRun parent;
    public int cycleCount;
    public Date startedAt, endedAt;
    public Grid beforeGrid, afterGrid;

    public GameCycle(GameRun parent) {
      this.parent = parent;
    }

    @Override
    public String toString() {
      return "GameCycle[cycle=" + cycleCount + ", " + "startedAt=" + startedAt + ", endedAt=" + endedAt + "]";
    }
  }

}
