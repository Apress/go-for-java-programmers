package org.baf.gol;

import java.util.Date;
import java.util.List;
import java.util.Map;

/**
 * A simple (but restricted) JSON object formatter.
 */
public class JsonFormatter implements Formatter {
  boolean pretty;
  String eol;

  public JsonFormatter(boolean pretty) {
    this.pretty = pretty;
    this.eol = pretty ? "\n" : "";
  }

  public JsonFormatter() {
    this(true);
  }

  @Override
  public String toString() {
    return "JsonFormatter[pretty=" + pretty + "]";
  }

  @Override
  public String valueToText(Object v) {
    StringBuilder sb = new StringBuilder();
    var size = 0;
    if (v instanceof List) {
      size = ((List) v).size();
    } else if (v instanceof Map) {
      size = ((Map) v).size();
    }
    valueToText(v, 0, "  ", "", size, ",  ", sb);
    return sb.toString();
  }

  // Format worker.
  void valueToText(Object v, int depth, String indent, String label, int len, String join, StringBuilder out) {
    if (join == null) {
      join = ", ";
    }
    var xindent = indent.repeat(depth);
    out.append(xindent);
    if (!label.isEmpty()) {
      out.append(label);
      out.append(": ");
    }
    if (v == null) {
      out.append("null");
      return;
    }
    // treat all implementations the same
    var c = v.getClass();
    var cname = c.getName();
    if (v instanceof List) {
      cname = List.class.getName();
    } else if (v instanceof Map) {
      cname = Map.class.getName();
    }
    // process all supported embedded types
    switch (cname) {
      case "java.util.Date":
        out.append(((Date) v).getTime());
        break;
      case "java.lang.String":
        v = '"' + v.toString().replace("\"", "\\\"") + '"';
      case "java.lang.Byte":
      case "java.lang.Short":
      case "java.lang.Integer":
      case "java.lang.Long":
      case "java.lang.Double":
      case "java.lang.Float":
      case "java.lang.Boolean":
        out.append(v.toString());
        break;
      case "java.util.List":
        out.append("[\n");
        List list = (List) v;
        for (int i = 0, xc = list.size(); i < xc; i++) {
          valueToText(list.get(i), depth + 1, indent, "", xc, join, out);
          out.append(i < len - 1 ? join : "");
          out.append(eol);
        }
        out.append(xindent + "]");
        break;
      case "java.util.Map":
        out.append("{\n");
        Map map = (Map) v;
        int i = 0, xc = map.size();
        for (var k : map.keySet()) {
          valueToText(map.get(k), depth + 1, indent, "\"" + k + "\"", xc, join, out);
          out.append(i < len - 1 ? join : "");
          i++;
          out.append(eol);
        }
        out.append(xindent + "}");
        break;
      default:
        throw new IllegalArgumentException("unknown type: " + cname);
    }
  }
}