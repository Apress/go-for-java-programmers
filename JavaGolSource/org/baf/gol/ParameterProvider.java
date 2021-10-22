package org.baf.gol;

/**
 * Provides a selected set of parameter values.
 */
public interface ParameterProvider {
  String getUrlParameter();

  String getNameParameter();

  int getMagFactorParameter();

  int getGameCyclesParameter();

  boolean startServerFlag();

  boolean runTimingsFlag();

  boolean reportFlag();

  boolean saveImageFlag();
}
