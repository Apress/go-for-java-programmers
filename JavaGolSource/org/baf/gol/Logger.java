package org.baf.gol;

import java.io.PrintStream;
import java.text.SimpleDateFormat;
import java.util.Date;

/**
 * Approximates the default Go logger function.
 *
 */
public class Logger {
  static public Logger log = new Logger();

  public PrintStream ps = System.out;
  public String lineFormat = "%-25s %-20s %-8s %-30s %s%n";
  public String contextFormat = "%s#%s@%d";
  public String threadFormat = "%s:%s";
  public SimpleDateFormat df = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSS");

  public void fatalf(String format, Object... args) {
    output(2, "FATAL", format, args);
    System.exit(3);
  }

  public void exceptionf(Exception e, String format, Object... args) {
    output(2, "EXCPT", "%s; caused by %s", String.format(format, args), e.getMessage());
    e.printStackTrace(ps);
  }

  public void errorf(String format, Object... args) {
    output(2, "ERROR", format, args);
  }

  public void tracef(String format, Object... args) {
    output(2, "TRACE", format, args);
  }

  void output(int level, String severity, String format, Object... args) {
    var text = String.format(format, args);
    Thread ct = Thread.currentThread();
    var st = ct.getStackTrace();
    StackTraceElement ste = st[level + 1];
    var tn = String.format(threadFormat, ct.getThreadGroup().getName(), ct.getName());
    var ctx = String.format(contextFormat, reduce(ste.getClassName()), ste.getMethodName(), ste.getLineNumber());
    ps.printf(lineFormat, df.format(new Date()), tn, severity, ctx, text);
  }

  String reduce(String name) {
    var posn = name.lastIndexOf(".");
    return posn >= 0 ? name.substring(posn + 1) : name;
  }
}
