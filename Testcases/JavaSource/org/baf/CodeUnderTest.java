package org.baf;

import java.math.BigInteger;
import java.util.Random;

public class CodeUnderTest {

  public static int echoInt(int in) {
    randomSleep(50);
    return in;
  }

  public static double echoFloat(double in) {
    randomSleep(50);
    return in;
  }

  public static String echoString(String in) {
    randomSleep(50);
    return in;
  }

  public static int sumInt(int in1, int in2) {
    randomSleep(50);
    return in1 + in2;
  }

  public static double sumFloat(double in1, double in2) {
    randomSleep(50);
    return in1 + in2;
  }

  public static String sumString(String in1, String in2) {
    randomSleep(50);
    return in1 + in2;
  }

//Factorial computation: factorial(n):
//n < 0 - undefined
//n == 0 - 1
//n > 0 - n * factorial(n-1)

  public static BigInteger factorialIterative(long n) {
    if (n < 0) {
      throw new IllegalArgumentException("invalid input");
    }
    BigInteger res = BigInteger.ONE;
    if (n == 0) {
      return res;
    }
    for (long i = 1; i <= n; i++) {
      res = res.multiply(new BigInteger(Long.toString(i)));
    }
    return res;
  }

  public static BigInteger factorialRecursive(long n) {
    if (n < 0) {
      throw new IllegalArgumentException("invalid input");
    }
    BigInteger res = BigInteger.ONE;
    if (n == 0) {
      return res;
    }
    // return new BigInteger(Long.toString(n)).multiply(factorialRecursive(n - 1));
    return factorialRecursive(n - 1).multiply(new BigInteger(Long.toString(n)));
  }

  private static Random rand = new Random();

  private static void randomSleep(int durMs) {
    try {
      Thread.sleep(1 + rand.nextInt(durMs));
    } catch (InterruptedException e) {
      // ignore
    }

  }

}
