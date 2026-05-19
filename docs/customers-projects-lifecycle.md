# Customers / Projects Lifecycle

Ponti keeps the domain relationship:

```txt
Actor    = identity master
Customer = operational customer
Project  = operational unit under customer
```

Rules:

- `customers.actor_id` links customer to the identity master.
- `projects.customer_id` remains the project relationship.
- Do not create `actor.projects`.
- Do not move projects directly to actors.

## Archive Customer

Archiving a customer:

- creates an archive batch with root `customers`;
- archives the customer;
- archives only active projects of that customer;
- marks those projects with the same archive batch and origin
  `customers/{customer_id}`;
- runs in one transaction.

## Restore Customer

Restoring a customer:

- restores the customer;
- restores only projects archived by the same customer archive batch;
- does not restore projects archived manually before;
- does not restore projects archived by another cause.

## Archive Project

Archiving a project:

- creates an archive batch with root `projects`;
- archives the project;
- archives project-owned rows according to project policy;
- does not archive the customer.

## Restore Project

Restoring a project:

- requires the customer to be active;
- restores only rows archived by that project archive batch;
- blocks if the customer is archived.

