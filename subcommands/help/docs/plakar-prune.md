PLAKAR-PRUNE(1) - General Commands Manual

# NAME

**plakar-prune** - Remove snapshots from a Plakar repository

# SYNOPSIS

**plakar&nbsp;prune**
\[*snapshotID&nbsp;...*]

# DESCRIPTION

The
**plakar prune**
command deletes snapshots from a Plakar repository.
Snapshots can be filtered for deletion by age, by tag, or by
specifying the snapshot IDs to remove.
If no
*snapshotID*
are provided, either
**-older**
or
**-tag**
must be specified to filter the snapshots to delete.

**plakar prune**
supports the location flags documented in
plakar-query(7)
to precisely select snapshots.

# EXAMPLES

Remove a specific snapshot by ID:

	$ plakar prune abc123

Remove snapshots older than 30 days:

	$ plakar prune -before 30d

Remove snapshots with a specific tag:

	$ plakar prune -tag daily-backup

Remove snapshots older than 1 year with a specific tag:

	$ plakar prune -before 1y -tag daily-backup

# DIAGNOSTICS

The **plakar-prune** utility exits&#160;0 on success, and&#160;&gt;0 if an error occurs.

0

> Command completed successfully.

&gt;0

> An error occurred, such as invalid date format or failure to delete a
> snapshot.

# SEE ALSO

plakar(1),
plakar-backup(1),
plakar-query(7)

Plakar - September 10, 2025
