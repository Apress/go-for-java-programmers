package org.baf.gol;

public class XmlFormatter implements Formatter {

  @Override
  public String valueToText(Object v) {
    throw new IllegalThreadStateException("not implemented");
  }
}

//boolean pretty;
//String eol;
//this.pretty = pretty;
//this.eol = pretty ? "\n" : "";
//@Override
//public String toString() {
//  return "XmlFormatter[pretty=" + pretty + "]";
//}
//StringBuilder sb = new StringBuilder();
//var size = 0;
//if (v instanceof List) {
//size = ((List) v).size();
//} else if (v instanceof Map) {
//size = ((Map) v).size();
//}
//appendOpen(sb, "data", null);
//valueToText(v, "", size, sb);
//appendClose(sb, "data");
//return sb.toString();
//static final char QUOTE = '"';
//
//void appendOpen(StringBuilder out, String label, Map<String, String> attrs) {
//  if (!label.isEmpty()) {
//    out.append("<");
//    out.append(label);
//    if (attrs != null && !attrs.isEmpty()) {
//      for (var k : attrs.keySet()) {
//        out.append(" ");
//        out.append(k);
//        out.append("=");
//        out.append('"' + attrs.get(k) + '"');
//      }
//    }
//    out.append(">");
//  }
//}
//
//void appendClose(StringBuilder out, String label) {
//  if (!label.isEmpty()) {
//    out.append("</");
//    out.append(label);
//    out.append(">");
//  }
//}
//
//void valueToText(Object v, String label, int len, StringBuilder out) {
//  if (v == null) {
//    appendOpen(out, label, null);
//    appendClose(out, label);
//    return;
//  }
//  var c = v.getClass();
//  var cname = c.getName();
//  if (v instanceof List) {
//    cname = List.class.getName();
//  } else if (v instanceof Map) {
//    cname = Map.class.getName();
//  }
//  switch (cname) {
//    case "java.util.Date":
//      out.append(((Date) v).getTime());
//      break;
//    case "java.lang.String":
//      v = QUOTE + v.toString().replace("\"", "\\\"") + QUOTE;
//    case "java.lang.Byte":
//    case "java.lang.Short":
//    case "java.lang.Integer":
//    case "java.lang.Long":
//    case "java.lang.Double":
//    case "java.lang.Float":
//    case "java.lang.Boolean": {
//      var attrs = new TreeMap<String, String>();
//      attrs.put("type", cname);
//      appendOpen(out, label, attrs);
//      out.append(v.toString());
//      appendClose(out, label);
//    }
//      break;
//    case "java.util.List": {
//      List list = (List) v;
//      for (int i = 0, xc = list.size(); i < xc; i++) {
//        var attrs = new TreeMap<String, String>();
//        attrs.put("type", cname);
//        appendOpen(out, "e", attrs);
//        valueToText(list.get(i), "", xc, out);
//        appendClose(out, "e");
//      }
//    }
//      break;
//    case "java.util.Map": {
//      Map map = (Map) v;
//      int xc = map.size();
//      for (var k : map.keySet()) {
//        var attrs = new TreeMap<String, String>();
////        attrs.put("type", cname);
////        appendOpen(out, "map", attrs);
//        valueToText(map.get(k), (String) k, xc, out);
////        appendClose(out, "map");
//      }
//    }
//      break;
//    default:
//      throw new IllegalArgumentException("unknown type: " + cname);
//  }
//}
//
//// test case: do not put in book
//public static void main(String[] args) {
//  var xf = new XmlFormatter();
//
//  var list = List.of(1, 2, 3, 4);
//  var map = new TreeMap<String, Object>();
//  map.put("1", "1");
//  map.put("s", "s");
//  var map2 = new TreeMap<String, Object>();
//  map2.put("1", "1");
//  map2.put("s", "s");
//  map.put("map", map2);
//
//  var v = new LinkedHashMap<String, Object>();
//  v.put("xnull", null);
//  v.put("xone", 1);
//  v.put("xtwo", 2.01);
//  v.put("xbool", true);
//  v.put("xmap", map);
//  v.put("xlist", list);
//  System.out.printf("in: %s%n", v);
//
//  xf.valueToText(v);
//  System.out.printf("out: %s%n", xf.valueToText(v));
//}

//@Override
//public String valueToText(Object v) {
//  String result = null;
//  if (!(v instanceof Map)) {
//    throw new RuntimeException("not implemented");
//  }
//  try {
//    // JAXB was in Java 8; no longer standard part of JSE 10+;
//    // needed to add external JAR
//    var ctx = JAXBContext.newInstance(v.getClass());
//    var marshaller = ctx.createMarshaller();
//    marshaller.setProperty(Marshaller.JAXB_FORMATTED_OUTPUT, true);
//
//    var baos = new ByteArrayOutputStream();
//    marshaller.marshal(v, baos);
//    result = new String(baos.toByteArray(), "UTF-8");
//  } catch (Exception e) {
//    log.exceptionf(e, "valueToTextfailed");
//  }
//  return result;
//}
