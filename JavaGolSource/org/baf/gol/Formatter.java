package org.baf.gol;

/**
 * Define a formatter (object to text).
 */
@FunctionalInterface
public interface Formatter {

  String valueToText(Object v);

}