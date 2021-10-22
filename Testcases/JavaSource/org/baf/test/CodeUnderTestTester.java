package org.baf.test;

import static org.junit.jupiter.api.Assertions.fail;

import java.math.BigInteger;

import org.baf.CodeUnderTest;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

class CodeUnderTestTester {
  private static final int nanosPerMilli = 1_000_000;

  private static final String factorial100Expect = "93326215443944152681699238856266700490715968264381621468592963895217599993229915608941463976156518286253697920827223758251185210916864000000000000000000000000";

  static long start;
  static int limit = 10000;

  @Test
  void testEchoInt() {
    System.out.println("in testEchoInt");
    int expect = 10;
    int got = CodeUnderTest.echoInt(expect);
    if (got != expect) {
      reportNoMatch(got, expect);
    }
  }

  @Test
  void testEchoFloat() {
    System.out.println("in testEchoFloat");
    double expect = 10;
    double got = CodeUnderTest.echoFloat(expect);
    if (got != expect) {
      reportNoMatch(got, expect);
    }
  }

  @Test
  void testEchoString() {
    System.out.println("in testEchoString");
    String expect = "hello";
    String got = CodeUnderTest.echoString(expect);
    if (!got.equals(expect)) {
      reportNoMatch(got, expect);
    }
  }

  @Test
  void testFactorialIterate() {
    System.out.println("in testFactorialIterate");
    BigInteger expect = new BigInteger(factorial100Expect);
    BigInteger got = CodeUnderTest.factorialIterative(100);
    if (!got.equals(expect)) {
      reportNoMatch(got, expect);
    }
  }

  @Test
  void testFactorialRecurse() {
    System.out.println("in testFactorialRecurse");
    BigInteger expect = new BigInteger(factorial100Expect);
    BigInteger got = CodeUnderTest.factorialRecursive(100);
    if (!got.equals(expect)) {
      reportNoMatch(got, expect);
    }
  }

//  @Test
//  void benchmarkSumInt() {
//    System.out.println("in benchmarkSumInt");
//    long start = System.currentTimeMillis();
//    for (int i = 0; i < limit; i++) {
//      CodeUnderTest.sumInt(10, 10);
//    }
//    long end = System.currentTimeMillis(), delta = end - start;
//    System.out.printf("factorialIterativeve : iterations=%d, totalTime=%.2fs, per call=%dns%n", limit,
//        (double) delta / 1000, delta * nanosPerMilli / limit);
//  }
//
//  @Test
//  void benchmarkSumFloat() {
//    System.out.println("in benchmarkSumFloat");
//    long start = System.currentTimeMillis();
//    for (int i = 0; i < limit; i++) {
//      CodeUnderTest.sumFloat(10, 10);
//    }
//    long end = System.currentTimeMillis(), delta = end - start;
//    System.out.printf("benchmarkSumFloat : iterations=%d, totalTime=%.2fs, per call=%dns%n", limit,
//        (double) delta / 1000, delta * nanosPerMilli / limit);
//  }
//
//  @Test
//  void benchmarkSumString() {
//    System.out.println("in benchmarkSumString");
//    long start = System.currentTimeMillis();
//    for (int i = 0; i < limit; i++) {
//      CodeUnderTest.sumString("Hello ", "World!");
//    }
//    long end = System.currentTimeMillis(), delta = end - start;
//    System.out.printf("benchmarkSumString : iterations=%d, totalTime=%.2fs, per call=%dns%n", limit,
//        (double) delta / 1000, delta * nanosPerMilli / limit);
//  }

  @Test
  void benchmarkFactorialInt() {
    System.out.println("in benchmarkFactorialInt");
    long start = System.currentTimeMillis();
    for (int i = 0; i < limit; i++) {
      CodeUnderTest.factorialIterative(100);
    }
    long end = System.currentTimeMillis(), delta = end - start;
    System.out.printf("benchmarkFactorialInt : iterations=%d, totalTime=%.2fs, per call=%dns%n", limit,
        (double) delta / 1000, delta * nanosPerMilli / limit);
  }

  @Test
  void benchmarkFactorialRec() {
    System.out.println("in benchmarkFactorialRec");
    long start = System.currentTimeMillis();
    for (int i = 0; i < limit; i++) {
      CodeUnderTest.factorialRecursive(100);
    }
    long end = System.currentTimeMillis(), delta = end - start;
    System.out.printf("benchmarkFactorialRec : iterations=%d, totalTime=%.2fs, per call=%dns%n", limit,
        (double) delta / 1000, delta * nanosPerMilli / limit);
  }

  @BeforeAll
  static void setUp() throws Exception {
    System.out.printf("starting tests...%n");
    start = System.currentTimeMillis();
  }

  @AfterAll
  static void tearDown() throws Exception {
    long end = System.currentTimeMillis();
    System.out.printf("tests complete in %dms%n", end - start);
  }

  private void reportNoMatch(Object got, Object expect) {
    fail(String.format("got(%s) != expect(%s)", got.toString(), expect.toString()));
  }

  private void reportFail(String message) {
    fail(String.format("failure: %s", message));
  }
}
