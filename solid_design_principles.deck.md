---
title: SOLID Principles in Software Engineering
theme: tokyo-night
routes:
  quick: intro -> why-solid -> hub -> conclusion
  deep-dive: intro -> why-solid -> hub -> srp -> srp-violation -> srp-clean -> ocp -> ocp-violation -> ocp-clean -> lsp -> lsp-violation -> lsp-clean -> lsp-rules -> isp -> isp-violation -> isp-clean -> dip -> dip-violation -> dip-clean -> synergy -> conclusion
---

# SOLID Principles in Software Engineering {#intro}
::tags: architecture, oop, solid, best-practices

### Managing Complexity in Large-Scale Object-Oriented Systems
**Author:** Sachin Yadav · **Format:** Interactive Technical Masterclass

As systems grow, classes and objects multiply exponentially:
- How do we manage complicated relationships and entities?
- How do we keep systems loosely coupled so adding features remains easy?
- How do we write clean code that new engineers understand immediately?

> Press Space or Enter to begin · Press : or Ctrl+P for Command Palette

---

# Why SOLID? Problems in Software Design {#why-solid}
::tags: problem-statement, maintenance

Every growing software project faces three recurring friction points:

1. **Maintainability**: Adding new features shouldn't require rewrites or break existing code.
2. **Readability**: Onboarding new engineers should be fast; code must state its intent clearly.
3. **Bug Isolation**: Fixing a defect in one subsystem should never cause regressions elsewhere.

### The SOLID Solution
Formulated by Robert C. Martin ("Uncle Bob"), the SOLID principles provide an architectural compass:

- **S**: Single Responsibility Principle (SRP)
- **O**: Open-Closed Principle (OCP)
- **L**: Liskov Substitution Principle (LSP)
- **I**: Interface Segregation Principle (ISP)
- **D**: Dependency Inversion Principle (DIP)

---

# Decision Hub: Choose Your Principle {#hub}
::tags: navigation, branching, roadmap

Jump directly into any principle or follow sequentially:

::branch [1] S — Single Responsibility Principle (SRP) -> srp
::branch [2] O — Open-Closed Principle (OCP) -> ocp
::branch [3] L — Liskov Substitution Principle (LSP) -> lsp
::branch [4] I — Interface Segregation Principle (ISP) -> isp
::branch [5] D — Dependency Inversion Principle (DIP) -> dip
::branch [6] Summary & Final Takeaways -> conclusion

> Tip: Press 1-5 to branch · Press Backspace or U to return to this fork anytime.

---

# Single Responsibility Principle (SRP) {#srp}
::tags: srp, fundamentals

### "A class should have one, and only one, reason to change."

- **1 Class = 1 Responsibility**
- Every attribute and method in a class should contribute to a single cohesive purpose.
- **Analogy**: A TV remote controls the TV—it shouldn't also regulate your refrigerator or start your car!

>[!NOTE]
> Single Responsibility does **not** mean one class has only one method.
> A class can have many helper and worker methods, as long as they all serve the exact same responsibility.

::branch [1] View SRP Code Violation -> srp-violation
::branch [2] View Clean SRP Solution -> srp-clean
::branch [3] Return to Decision Hub -> hub

---

# SRP Violation: Multi-Responsibility Cart {#srp-violation}
::tags: srp, anti-pattern, cpp

Here, `ShoppingCart` violates SRP by mixing calculation, invoice printing, and database persistence:

```cpp
// VIOLATION: ShoppingCart has 3 different reasons to change!
class ShoppingCart {
private:
    vector<Product*> products;
public:
    void addProduct(Product* p) { products.push_back(p); }
    const vector<Product*>& getProducts() { return products; }

    // Responsibility 1: Cart business logic
    double calculateTotal() {
        double total = 0;
        for (auto p : products) total += p->price;
        return total;
    }

    // Responsibility 2: Presentation & formatting (Violates SRP)
    void printInvoice() {
        cout << "Shopping Cart Invoice:\n";
        for (auto p : products) cout << p->name << " - $" << p->price << endl;
        cout << "Total: $" << calculateTotal() << endl;
    }

    // Responsibility 3: Database storage (Violates SRP)
    void saveToDatabase() {
        cout << "Saving shopping cart to database..." << endl;
    }
};
```

---

# SRP Applied: Cohesive Separation {#srp-clean}
::tags: srp, solution, cpp

Separate distinct concerns into dedicated, single-purpose classes:

```cpp
// 1. ShoppingCart: Business logic only
class ShoppingCart {
private:
    vector<Product*> products;
public:
    void addProduct(Product* p) { products.push_back(p); }
    const vector<Product*>& getProducts() { return products; }
    double calculateTotal() {
        double total = 0;
        for (auto p : products) total += p->price;
        return total;
    }
};

// 2. ShoppingCartPrinter: Presentation only
class ShoppingCartPrinter {
public:
    static void print(ShoppingCart* cart) {
        for (auto p : cart->getProducts()) cout << p->name << " - $" << p->price << endl;
        cout << "Total: $" << cart->calculateTotal() << endl;
    }
};

// 3. ShoppingCartStorage: Persistence only
class ShoppingCartStorage {
public:
    static void save(ShoppingCart* cart) {
        cout << "Saving shopping cart to database..." << endl;
    }
};
```

---

# Open-Closed Principle (OCP) {#ocp}
::tags: ocp, fundamentals

### "Software entities should be open for extension, but closed for modification."

- **Open for Extension**: You can introduce new features and behaviors seamlessly.
- **Closed for Modification**: You do **not** touch or alter tested, production code to add new behavior.

>[!IMPORTANT]
> How do we achieve this without modifying existing files?
> **Abstraction + Polymorphism**. We define stable interfaces and implement new behavior in new concrete classes!

::branch [1] View OCP Code Violation -> ocp-violation
::branch [2] View Clean OCP Solution -> ocp-clean
::branch [3] Return to Decision Hub -> hub

---

# OCP Violation: Modifying Existing Code {#ocp-violation}
::tags: ocp, anti-pattern, cpp

When requirements change to support MongoDB and Flat Files, directly editing `ShoppingCartStorage` violates OCP:

```cpp
// VIOLATION: Adding new storage types modifies existing tested class
class ShoppingCartStorage {
private:
    ShoppingCart* cart;
public:
    ShoppingCartStorage(ShoppingCart* cart) : cart(cart) {}

    void saveToSQLDatabase() {
        cout << "Saving shopping cart to SQL DB..." << endl;
    }

    // Direct modification: touches old class to add Mongo
    void saveToMongoDatabase() {
        cout << "Saving shopping cart to Mongo DB..." << endl;
    }

    // Direct modification: touches old class to add File
    void saveToFile() {
        cout << "Saving shopping cart to File..." << endl;
    }
};
```

---

# OCP Applied: Persistence Abstraction {#ocp-clean}
::tags: ocp, solution, cpp

Declare an abstract `Persistence` interface. Adding a new database requires writing a new class, with zero changes to existing classes:

```cpp
// Stable Interface (Closed for modification)
class Persistence {
public:
    virtual void save(ShoppingCart* cart) = 0;
    virtual ~Persistence() = default;
};

// Open for extension: Concrete implementations
class SQLPersistence : public Persistence {
public:
    void save(ShoppingCart* cart) override {
        cout << "Saving shopping cart to SQL DB..." << endl;
    }
};

class MongoPersistence : public Persistence {
public:
    void save(ShoppingCart* cart) override {
        cout << "Saving shopping cart to MongoDB..." << endl;
    }
};

class FilePersistence : public Persistence {
public:
    void save(ShoppingCart* cart) override {
        cout << "Saving shopping cart to a file..." << endl;
    }
};
```

---

# Liskov Substitution Principle (LSP) {#lsp}
::tags: lsp, fundamentals

### "Subclasses must be substitutable for their base classes."

If class `B` is a subtype of class `A`, any client expecting `A` must function correctly when given `B` without surprise exceptions or type checks.

- A subclass should **only expand** on a base class, never narrow or break its contract.
- If a method override throws `NotSupportedException` or `logic_error`, LSP is broken!

::branch [1] View LSP Bank Account Violation -> lsp-violation
::branch [2] View The Flawed "Conditional Fix" -> lsp-bad-fix
::branch [3] View Clean LSP Hierarchy -> lsp-clean
::branch [4] Deep Dive: The 4 Core LSP Rules -> lsp-rules

---

# LSP Violation: Breaking Subtypes {#lsp-violation}
::tags: lsp, anti-pattern, cpp

`FixedTermAccount` inherits from `Account`, but throws an exception on `withdraw()`:

```cpp
class Account {
public:
    virtual void deposit(double amount) = 0;
    virtual void withdraw(double amount) = 0;
};

class SavingAccount : public Account { /* allows deposit & withdraw */ };
class CurrentAccount : public Account { /* allows deposit & withdraw */ };

// VIOLATION: FixedTermAccount cannot be substituted for Account!
class FixedTermAccount : public Account {
public:
    void deposit(double amount) override { /* ok */ }

    void withdraw(double amount) override {
        // Crashes client expecting normal Account behavior!
        throw logic_error("Withdrawal not allowed in Fixed Term Account!");
    }
};
```

---

# LSP: The Flawed "Conditional Fix" {#lsp-bad-fix}
::tags: lsp, anti-pattern, cpp

### The Wrong Way to Fix LSP: Client-Side Type Checking

```cpp
// BAD APPROACH: Client uses runtime type inspection
void processTransactions(vector<Account*> accounts) {
    for (Account* acc : accounts) {
        acc->deposit(1000);

        // Anti-pattern: Client couples directly to concrete types!
        if (dynamic_cast<FixedTermAccount*>(acc) == nullptr) {
            acc->withdraw(500);
        }
    }
}
```

>[!CAUTION]
> Type-checking breaks **both** LSP and OCP!
> Every time a new account type is created, every client loop must be inspected and updated.

---

# LSP Applied: Segregated Hierarchy {#lsp-clean}
::tags: lsp, solution, cpp

Separate the account contracts into `DepositOnlyAccount` and `WithdrawableAccount`:

```cpp
class DepositOnlyAccount {
public:
    virtual void deposit(double amount) = 0;
    virtual ~DepositOnlyAccount() = default;
};

class WithdrawableAccount : public DepositOnlyAccount {
public:
    virtual void withdraw(double amount) = 0;
};

class SavingAccount : public WithdrawableAccount { /* implements deposit & withdraw */ };
class CurrentAccount : public WithdrawableAccount { /* implements deposit & withdraw */ };

// Clean: FixedTermAccount only implements DepositOnlyAccount!
class FixedTermAccount : public DepositOnlyAccount {
public:
    void deposit(double amount) override { cout << "Deposited into Fixed Term Account.\n"; }
};
```

Clients receive `vector<WithdrawableAccount*>` and `vector<DepositOnlyAccount*>`. Zero exceptions, zero type checking.

---

# LSP Deep-Dive: The 4 Rule Categories {#lsp-rules}
::tags: lsp, advanced, rules

To ensure subclasses are truly substitutable, adhere to four rule categories:

```
                  ┌────────────────────────────────────────┐
                  │    Liskov Substitution Rules (LSP)     │
                  └───────────────────┬────────────────────┘
          ┌───────────────────┬───────┴──────────┬───────────────────┐
          ▼                   ▼                  ▼                   ▼
    1. Signature Rules   2. Exceptions      3. Properties       4. Methods
    • Contravariant Args • Fewer/narrower   • Class Invariants  • Weaken Preconds
    • Covariant Returns    exceptions         maintained        • Strengthen Post
```

::branch [1] Rule 1: Signature & Covariance -> lsp-sig
::branch [2] Rule 2: Exception Hierarchy -> lsp-exc
::branch [3] Rule 3: Class Invariants & History -> lsp-prop
::branch [4] Rule 4: Pre & Post Conditions -> lsp-cond
::branch [5] Continue to Interface Segregation (ISP) -> isp

---

# LSP Rule 1: Signature & Covariance {#lsp-sig}
::tags: lsp, rules, signatures

### Argument Contravariance & Return Type Covariance

1. **Method Arguments**: Subclass parameter types must be identical or broader than superclass types (C++ enforces identical signatures).
2. **Return Type Covariance**: Subclass return types can be identical or **narrower** (more specific), never broader:

```cpp
class Animal { /* base */ };
class Dog : public Animal { /* narrower derived */ };

class Parent {
public:
    virtual Animal* getAnimal() { return new Animal(); }
};

class Child : public Parent {
public:
    // Covariant return type: Returning Dog* satisfies Animal* contract!
    Dog* getAnimal() override { return new Dog(); }
};
```

---

# LSP Rule 2: Exception Rules {#lsp-exc}
::tags: lsp, rules, exceptions

### Subclasses must throw fewer or narrower exceptions

Clients catch exceptions based on the base class contract. If a subclass throws an unfamiliar or broader exception, the client fails:

```cpp
// Exception hierarchy:
// std::logic_error
//   └── std::out_of_range

class Parent {
public:
    virtual void getValue() { throw logic_error("Parent error"); }
};

class Child : public Parent {
public:
    // Narrower exception: Caught by catch(const logic_error&) -> LSP Compliant
    void getValue() override { throw out_of_range("Child range error"); }

    // WRONG: Throwing std::runtime_error breaks client contract!
};
```

---

# LSP Rule 3: Invariants & History {#lsp-prop}
::tags: lsp, rules, invariants

### Property Rules: Invariants & History Constraints

1. **Class Invariant**: A condition that remains true for an object throughout its entire lifetime.
   - Example: `BankAccount` balance can never be negative.
   - Subclasses must **maintain or strengthen** the invariant, never weaken it.
2. **History Constraint**: State changes disallowed by the base class must never be allowed by subclasses:

```cpp
class BankAccount {
protected:
    double balance;
public:
    virtual void withdraw(double amount) {
        if (balance - amount < 0) throw runtime_error("Insufficient funds");
        balance -= amount;
    }
};

// VIOLATION: CheatAccount breaks invariant by allowing negative balance
class CheatAccount : public BankAccount {
public:
    void withdraw(double amount) override {
        balance -= amount; // LSP Violation: Invariant broken!
    }
};
```

---

# LSP Rule 4: Pre- & Post-Conditions {#lsp-cond}
::tags: lsp, rules, conditions

### Preconditions (Weakened) & Postconditions (Strengthened)

- **Precondition**: Requirement before method runs. Subclasses can **weaken** it (be more lenient), but never strengthen it.
  - *Example*: Base requires password >= 8 chars; Subclass accepts password >= 6 chars (weakened: OK).
- **Postcondition**: Guarantee after method finishes. Subclasses can **strengthen** it (guarantee more), but never weaken it:

```cpp
class Car {
public:
    // Postcondition: Speed reduces after braking
    virtual void brake() { speed -= 20; }
};

class HybridCar : public Car {
public:
    // Strengthened Postcondition: Speed reduces AND battery recharges
    void brake() override {
        speed -= 20;
        batteryCharge += 10; // Extra guarantee: LSP Compliant!
    }
};
```

---

# Interface Segregation Principle (ISP) {#isp}
::tags: isp, fundamentals

### "Clients should not be forced to depend on methods they do not use."

- Many small, client-specific interfaces are better than one bloated general-purpose interface.
- Prevent "fat interfaces" that force implementers to write stub methods or throw `NotImplementedException`.

>[!NOTE]
> When an interface contains methods that only 50% of implementers actually need, segregate it into smaller modular interfaces.

::branch [1] View ISP Shape Violation -> isp-violation
::branch [2] View Clean ISP Segregation -> isp-clean
::branch [3] Return to Decision Hub -> hub

---

# ISP Violation: The Fat Interface {#isp-violation}
::tags: isp, anti-pattern, cpp

Forcing 2D shapes (`Square`, `Rectangle`) to implement 3D methods (`volume()`):

```cpp
// FAT INTERFACE: Violates Interface Segregation
class Shape {
public:
    virtual double area() = 0;
    virtual double volume() = 0; // 2D shapes do NOT have volume!
};

class Square : public Shape {
    double side;
public:
    Square(double s) : side(s) {}
    double area() override { return side * side; }

    // VIOLATION: Forced to implement useless method
    double volume() override {
        throw logic_error("Volume not applicable for Square");
    }
};
```

---

# ISP Applied: Segregated Interfaces {#isp-clean}
::tags: isp, solution, cpp

Decompose into focused `TwoDimensionalShape` and `ThreeDimensionalShape` interfaces:

```cpp
// Interface 1: 2D Shapes only
class TwoDimensionalShape {
public:
    virtual double area() = 0;
    virtual ~TwoDimensionalShape() = default;
};

// Interface 2: 3D Shapes
class ThreeDimensionalShape : public TwoDimensionalShape {
public:
    virtual double volume() = 0;
};

// Clean: Square only implements what it actually supports
class Square : public TwoDimensionalShape {
    double side;
public:
    Square(double s) : side(s) {}
    double area() override { return side * side; }
};

// Clean: Cube implements both area and volume
class Cube : public ThreeDimensionalShape {
    double side;
public:
    Cube(double s) : side(s) {}
    double area() override { return 6 * side * side; }
    double volume() override { return side * side * side; }
};
```

---

# Dependency Inversion Principle (DIP) {#dip}
::tags: dip, fundamentals

### "High-level modules should not depend on low-level modules. Both should depend on abstractions."

- **High-level modules**: Core business logic (e.g. `UserService`, `CheckoutService`).
- **Low-level modules**: Technical plumbing and infrastructure (e.g. `MySQLDatabase`, `EmailService`).
- Direct dependencies couple business logic to volatile implementation details.

>[!TIP]
> **Key Insight**: If the Open-Closed Principle is the goal, the Dependency Inversion Principle is the primary mechanism to achieve it!

::branch [1] View DIP Tight Coupling Violation -> dip-violation
::branch [2] View Clean Dependency Injection -> dip-clean
::branch [3] Return to Decision Hub -> hub

---

# DIP Violation: Tightly Coupled Service {#dip-violation}
::tags: dip, anti-pattern, cpp

`UserService` directly instantiates concrete `MySQLDatabase` and `MongoDBDatabase`:

```cpp
class MySQLDatabase {
public:
    void saveToSQL(string data) { cout << "INSERT INTO users VALUES('" << data << "')\n"; }
};

class MongoDBDatabase {
public:
    void saveToMongo(string data) { cout << "db.users.insert({name: '" << data << "'})\n"; }
};

// VIOLATION: High-level UserService is tightly bound to low-level DB engines
class UserService {
private:
    MySQLDatabase sqlDb;      // Direct dependency
    MongoDBDatabase mongoDb;  // Direct dependency
public:
    void storeUserToSQL(string user) { sqlDb.saveToSQL(user); }
    void storeUserToMongo(string user) { mongoDb.saveToMongo(user); }
};
```

---

# DIP Applied: Dependency Injection {#dip-clean}
::tags: dip, solution, cpp

Invert the dependency: `UserService` depends on the `Database` abstraction injected via constructor:

```cpp
// Abstraction (Interface)
class Database {
public:
    virtual void save(string data) = 0;
    virtual ~Database() = default;
};

// Low-level modules implement the abstraction
class MySQLDatabase : public Database {
public:
    void save(string data) override { cout << "INSERT INTO users VALUES('" << data << "')\n"; }
};

class MongoDBDatabase : public Database {
public:
    void save(string data) override { cout << "db.users.insert({name: '" << data << "'})\n"; }
};

// High-level module depends ONLY on the abstraction
class UserService {
private:
    Database* db; // Injected abstraction
public:
    UserService(Database* database) : db(database) {}
    void storeUser(string user) { db->save(user); }
};
```

---

# Architectural Synergy: OCP & DIP {#synergy}
::tags: architecture, summary, synergy

### How the SOLID Principles Reinforce Each Other

```
┌──────────────────────────────────────────────────────────────┐
│                    Clean Architecture Loop                   │
└──────────────────────────────┬───────────────────────────────┘
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
       SRP Isolates Reasons            DIP Decouples Modules
           to Change                       via Abstractions
               │                               │
               └───────────────┬───────────────┘
                               ▼
                    OCP Enables Extension
                     Without Modification
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
      LSP Preserves Subtype           ISP Keeps Interfaces
           Contracts                     Lean & Focused
```

- When you practice **SRP**, classes become small and cohesive.
- When you apply **ISP**, interfaces stay focused.
- When you implement **DIP**, dependencies flow toward interfaces.
- The result: **OCP** and **LSP** emerge naturally!

---

# Summary & Practical Takeaways {#conclusion}
::tags: summary, checklist

### The SOLID Mental Checklist

1. **S (Single Responsibility)**: Does this class have more than one reason to change?
2. **O (Open-Closed)**: Can I add this new feature without editing existing classes?
3. **L (Liskov Substitution)**: Can every subclass substitute for its parent without throwing surprise errors?
4. **I (Interface Segregation)**: Are clients forced to implement methods they don't care about?
5. **D (Dependency Inversion)**: Am I depending on concrete classes or stable interfaces?

>[!TIP]
> SOLID principles are architectural guides, not dogmatic laws.
> Apply them to manage complexity where change is expected, keeping your codebase extensible and maintainable.

Happy Architecting!
