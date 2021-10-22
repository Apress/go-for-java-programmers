package org.baf.gol;

import java.awt.image.BufferedImage;
import java.awt.image.RenderedImage;
import java.io.IOException;

import javax.imageio.IIOImage;
import javax.imageio.ImageIO;
import javax.imageio.ImageTypeSpecifier;
import javax.imageio.ImageWriteParam;
import javax.imageio.ImageWriter;
import javax.imageio.metadata.IIOMetadata;
import javax.imageio.metadata.IIOMetadataNode;
import javax.imageio.stream.ImageOutputStream;

/**
 * Supports combining multiple images into a single animated GIF.
 *
 */
public class AnnimatedGifWriter implements java.io.Closeable {
  private static final String CODE = "2.0";
  private static final String ID = "NETSCAPE";
  private static final String ZERO_INDEX = "0";
  private static final String NONE = "none";
  private static final String FALSE = "FALSE";

  protected IIOMetadata metadata;
  protected ImageWriter writer;
  protected ImageWriteParam params;

  public AnnimatedGifWriter(ImageOutputStream ios, int imageType, boolean showAsLoop, int delayMs, String author)
      throws IOException {
    var imageTypeSpecifier = ImageTypeSpecifier.createFromBufferedImageType(imageType);
    writer = ImageIO.getImageWritersBySuffix("gif").next();
    params = writer.getDefaultWriteParam();
    metadata = writer.getDefaultImageMetadata(imageTypeSpecifier, params);
    configMetadata(delayMs, showAsLoop, "Author: " + author);

    writer.setOutput(ios);
    writer.prepareWriteSequence(null);
  }

  @Override
  public void close() throws IOException {
    writer.endWriteSequence();
  }

  /**
   * Creates an animated GIF from 1+ images.
   */
  public static void createGifs(ImageOutputStream ios, int delay, String author, BufferedImage... images)
      throws IOException {
    if (delay < 0) {
      delay = 5 * 1000;
    }
    if (images.length < 1) {
      throw new IllegalArgumentException("at least one image is required");
    }
    try (var writer = new AnnimatedGifWriter(ios, images[0].getType(), true, delay, author)) {
      for (var image : images) {
        writer.addImage(image);
      }
    }
  }

  // configure self
  void configMetadata(int delay, boolean loop, String comment) throws IOException {
    var name = metadata.getNativeMetadataFormatName();
    var root = (IIOMetadataNode) metadata.getAsTree(name);
    metadata.setFromTree(name, root);

    var cel = findOrAddMetadata(root, "CommentExtensions");
    cel.setAttribute("CommentExtension", comment);

    var gce = findOrAddMetadata(root, "GraphicControlExtension");
    gce.setAttribute("transparentColorIndex", ZERO_INDEX);
    gce.setAttribute("userInputFlag", FALSE);
    gce.setAttribute("transparentColorFlag", FALSE);
    gce.setAttribute("delayTime", Integer.toString(delay / 10));
    gce.setAttribute("disposalMethod", NONE);

    byte[] bytes = new byte[] { 1, (byte) (loop ? 0 : 1), 0 };
    var ael = findOrAddMetadata(root, "ApplicationExtensions");
    var ae = new IIOMetadataNode("ApplicationExtension");
    ae.setUserObject(bytes);
    ae.setAttribute("authenticationCode", CODE);
    ae.setAttribute("applicationID", ID);
    ael.appendChild(ae);
  }

  static IIOMetadataNode findOrAddMetadata(IIOMetadataNode root, String metadataType) {
    for (int i = 0, c = root.getLength(); i < c; i++) {
      if (root.item(i).getNodeName().equalsIgnoreCase(metadataType)) {
        return (IIOMetadataNode) root.item(i);
      }
    }
    var node = new IIOMetadataNode(metadataType);
    root.appendChild(node);
    return (node);
  }

  void addImage(RenderedImage img) throws IOException {
    writer.writeToSequence(new IIOImage(img, null, metadata), params);
  }
}
