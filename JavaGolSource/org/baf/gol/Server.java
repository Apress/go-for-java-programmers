package org.baf.gol;

import static org.baf.gol.Logger.log;
import static org.baf.gol.Utility.NANOS_PER_MS;
import static org.baf.gol.Utility.isNullOrEmpty;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

/**
 * Provides a HTTP server for the GoL.<br>
 * Uses com.sun.net.httpserver.HttpServer for basic function.<br>
 * Can be opened only one time.
 **/
public class Server implements AutoCloseable {
  private static final String GIF_IMAGE_FILE_PATTERN = "/Image_%s.gif";

  String address;
  int port;
  Map<String, HttpHandler> handlers = new LinkedHashMap<>();
  HttpServer server;
  ParameterProvider pp;
  public String saveImageRoot = "/temp"; // change per OS type

  public Server(ParameterProvider pp) {
    this(pp, "localhost", 8080);
  }

  public Server(ParameterProvider pp, String address, int port) {
    this.pp = pp;
    this.address = address;
    this.port = port;
  }

  @Override
  public String toString() {
    return "Server[address=" + address + ", port=" + port + ", open=" + isOpen() + ", handlers=" + handlers.keySet()
        + "]";
  }

  String getRequestPath(HttpExchange ex) {
    return ex.getRequestURI().toString().split("\\?")[0];
  }

  // assumes only one value; redo if more than one possible
  String getQueryParamValue(HttpExchange ex, String name) {
    String result = null;
    var parts = ex.getRequestURI().toString().split("\\?");
    if (parts.length > 1) {
      parts = parts[1].split("&");
      for (var part : parts) {
        var xparts = part.split("=");
        if (xparts[0].equals(name)) {
          result = xparts[1];
          break;
        }
      }
    }
    return result;
  }

  /**
   * Used to allow clients outside this class to send data.
   */
  public interface ResponseDataSender {
    void sendResponseData(byte[] data) throws IOException;
  }

  public class DefaultResponseDataSender implements ResponseDataSender {
    HttpExchange ex;

    public DefaultResponseDataSender(HttpExchange ex) {
      this.ex = ex;
    }

    @Override
    public void sendResponseData(byte[] data) throws IOException {
      Server.this.sendResponseData(ex, data);
    }

  }

  void sendResponseData(HttpExchange ex, byte[] data) throws IOException {
    ex.sendResponseHeaders(200, data.length);
    var os = ex.getResponseBody();
    os.write(data);
    os.flush();
    os.close();
    log.tracef("Sent %d bytes", data.length);
  }

  void sendResponseJson(HttpExchange ex, Object data) throws IOException {
    ex.getResponseHeaders().add("Content-Type", "text/json");
    var jf = new JsonFormatter();
    sendResponseData(ex, jf.valueToText(data).getBytes());
  }

  void sendResponseXml(HttpExchange ex, Object data) throws IOException {
    ex.getResponseHeaders().add("Content-Type", "text/xml");
    var xf = new XmlFormatter();
    sendResponseData(ex, xf.valueToText(data).getBytes());
  }

  void sendStatus(HttpExchange ex, int status) throws IOException {
    ex.sendResponseHeaders(status, 0);
  }

// Show request handler.
  HttpHandler showHandler = new HttpHandler() {

    @Override
    public void handle(HttpExchange exchange) throws IOException {
      try {
        switch (exchange.getRequestMethod()) {
          case "GET": {
            if (!Objects.equals(getRequestPath(exchange), "/show")) {
              sendStatus(exchange, 404);
              return;
            }
            // process query parameters
            var name = getQueryParamValue(exchange, "name");
            if (isNullOrEmpty(name)) {
              name = "default";
            }
            var form = getQueryParamValue(exchange, "form");
            if (isNullOrEmpty(form)) {
              form = "gif";
            }
            var xmaxCount = getQueryParamValue(exchange, "maxCount");
            if (isNullOrEmpty(xmaxCount)) {
              xmaxCount = "50";
            }
            var maxCount = Integer.parseInt(xmaxCount);
            if (maxCount < 1 || maxCount > 100) {
              sendStatus(exchange, 400);
              return;
            }
            var xmag = getQueryParamValue(exchange, "mag");
            if (isNullOrEmpty(xmag)) {
              xmag = "1";
            }
            var mag = Integer.parseInt(xmag);
            var xindex = getQueryParamValue(exchange, "index");
            if (isNullOrEmpty(xindex)) {
              xindex = "0";
            }
            var index = Integer.parseInt(xindex);
            if (index < 0) {
              sendStatus(exchange, 400);
              return;
            }

            // get a game
            var gr = Game.coreGame.runs.get(name);
            if (gr == null) {
              sendStatus(exchange, 404);
              return;
            }

            // return requested image type
            switch (form) {
              case "GIF":
              case "gif": {
                var b = gr.makeGifs(maxCount, mag);
                sendResponseData(exchange, b);

                if (pp.saveImageFlag()) {
                  var imageFormat = saveImageRoot + GIF_IMAGE_FILE_PATTERN;
                  var saveFile = String.format(imageFormat, name);
                  Files.write(Paths.get(saveFile), b);
                  log.tracef("Save %s", saveFile);
                }
              }
                break;
              case "PNG":
              case "png": {
                if (index <= maxCount) {
                  var rs = new DefaultResponseDataSender(exchange);
                  gr.sendPng(rs, index, mag);
                } else {
                  sendStatus(exchange, 400);
                }
              }
                break;
              default:
                sendStatus(exchange, 405);
            }
          }
        }
      } catch (Exception e) {
        log.exceptionf(e, "show failed");
        sendStatus(exchange, 500);
      }
    }
  };

// Play request handler.
  HttpHandler playHandler = new HttpHandler() {

    @Override
    public void handle(HttpExchange exchange) throws IOException {
      try {
        switch (exchange.getRequestMethod()) {
          case "GET": {
            if (!Objects.equals(getRequestPath(exchange), "/play")) {
              sendStatus(exchange, 404);
              return;
            }
            // process query parameters
            var name = getQueryParamValue(exchange, "name");
            var url = getQueryParamValue(exchange, "url");
            if (Utility.isNullOrEmpty(name) || Utility.isNullOrEmpty(url)) {
              sendStatus(exchange, 400);
              return;
            }
            var ct = getQueryParamValue(exchange, "ct");
            if (Utility.isNullOrEmpty(ct)) {
              ct = exchange.getRequestHeaders().getFirst("Content-Type");
            }
            if (Utility.isNullOrEmpty(ct)) {
              ct = "";
            }
            ct = ct.toLowerCase();
            switch (ct) {
              case "":
                ct = "application/json";
                break;
              case "application/json":
              case "text/json":
                break;
              case "application/xml":
              case "text/xml":
                break;
              default:
                sendStatus(exchange, 400);
            }

            // run a game
            Game.coreGame.run(name, url);
            var run = makeReturnedRun(name, url);

            // return statistics as requested
            switch (ct) {
              case "application/json":
              case "text/json": {
                sendResponseJson(exchange, run);
              }
                break;
              case "application/xml":
              case "text/xml": {
                sendResponseXml(exchange, run);
              }
                break;
            }
          }
            break;
          default:
            sendStatus(exchange, 405);
        }
      } catch (Exception e) {
        log.exceptionf(e, "play failed");
        sendStatus(exchange, 500);
      }
    }
  };

// History request handler.
  HttpHandler historyHandler = new HttpHandler() {

    @Override
    public void handle(HttpExchange exchange) throws IOException {
      try {
        switch (exchange.getRequestMethod()) {
          case "GET": {
            if (!Objects.equals(getRequestPath(exchange), "/history")) {
              sendStatus(exchange, 404);
              return;
            }
            // format history
            Map<String, Object> game = new LinkedHashMap<>();
            var runs = new LinkedHashMap<>();
            game.put("Runs", runs);
            var xruns = Game.coreGame.runs;
            for (var k : xruns.keySet()) {
              runs.put(k, makeReturnedRun(k, xruns.get(k).imageUrl));
            }
            sendResponseJson(exchange, game);
          }
            break;
          case "DELETE":
            if (!Objects.equals(getRequestPath(exchange), "/history")) { // more is bad
              sendStatus(exchange, 404);
              return;
            }
            // erase history
            Game.coreGame.clear();
            sendStatus(exchange, 204);
            break;
          default:
            sendStatus(exchange, 405);
        }
      } catch (Exception e) {
        log.exceptionf(e, "history failed");
        sendStatus(exchange, 500);
      }
    }
  };

  Map<String, Object> makeReturnedRun(String name, String imageUrl) {
    var xrun = new LinkedHashMap<String, Object>();
    var run = Game.coreGame.runs.get(name);
    if (run != null) {
      xrun.put("Name", run.name);
      xrun.put("ImageURL", run.imageUrl);
      xrun.put("PlayIndex", run.playIndex);
      xrun.put("DelayIn10ms", run.delayIn10ms);
      xrun.put("Height", run.height);
      xrun.put("Width", run.width);
      xrun.put("StartedAMst", run.startedAt);
      xrun.put("EndedAMst", run.endedAt);
      xrun.put("DurationMs", run.endedAt.getTime() - run.startedAt.getTime());
      var cycles = new ArrayList<Map<String, Object>>();
      xrun.put("Cycles", cycles);
      for (var r : run.cycles) {
        var xc = new LinkedHashMap<String, Object>();
        xc.put("StartedAtNs", r.startedAt.getTime() * NANOS_PER_MS);
        xc.put("EndedAtNs", r.endedAt.getTime() * NANOS_PER_MS);
        var duration = (r.endedAt.getTime() - r.startedAt.getTime()) * NANOS_PER_MS;
        xc.put("DurationNs", duration);
        xc.put("Cycle", r.cycleCount);
        xc.put("ThreadCount", Game.coreGame.threadCount);
        xc.put("MaxCount", Game.coreGame.maxCycles);
        cycles.add(xc);
      }

    }
    return xrun;
  }

  public void startHttpServer() throws IOException {
    registerContext("/play", playHandler);
    registerContext("/show", showHandler);
    registerContext("/history", historyHandler);
    open();
    log.tracef("Server %s:%d started", address, port);
  }

  public void open() throws IOException {
    if (isOpen()) {
      throw new IllegalStateException("already open");
    }
    server = HttpServer.create(new InetSocketAddress("localhost", 8080), 0);
    for (var path : handlers.keySet()) {
      server.createContext(path, handlers.get(path));
    }
    server.start();
    Runtime.getRuntime().addShutdownHook(new Thread(() -> {
      try {
        close();
      } catch (Exception e) {
        log.exceptionf(e, "shutdown failed");
      }
    }));
  }

  public boolean isOpen() {
    return server != null;
  }

  @Override
  public void close() throws Exception {
    if (isOpen()) {
      server.stop(60);
      server = null;
    }
  }

  public void registerContext(String path, HttpHandler handler) {
    if (handlers.containsKey(path)) {
      throw new IllegalArgumentException("path already exists: " + path);
    }
    handlers.put(path, handler);
  }

  public void removeContext(String path) {
    if (!handlers.containsKey(path)) {
      throw new IllegalArgumentException("unknown path: " + path);
    }
    handlers.remove(path);
  }

  public List<String> getContextPaths() {
    return handlers.keySet().stream().collect(Collectors.toUnmodifiableList());
  }
}
