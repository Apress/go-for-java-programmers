package org.baf.gol;

import static org.baf.gol.Logger.log;

import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.net.URL;

import javax.imageio.ImageIO;

public class Utility {
  public static final int NANOS_PER_MS = 1_000_000;
  public static final String FILE_PREFIX = "file:";

  public static boolean isNullOrEmpty(CharSequence cs) {
    return cs == null || cs.length() == 0;
  }

  public static boolean isNullOrEmptyTrim(String cs) {
    return cs == null || cs.trim().length() == 0;
  }

  public static BufferedImage loadImage(String url, String[] kind) throws IOException {
    BufferedImage bi = null;
    if (url.startsWith(FILE_PREFIX)) {
      String name = url.substring(FILE_PREFIX.length());
      log.tracef("loadImage %s; %s", url, name);
      bi = ImageIO.read(new File(name));
    } else {
      var xurl = new URL(url);
      bi = ImageIO.read(xurl);
    }
    var posn = url.lastIndexOf(".");
    kind[0] = posn >= 0 ? url.substring(posn + 1) : "gif";
    return bi;
  }

}
