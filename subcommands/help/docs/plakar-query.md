PLAKAR-QUERY(7) - Miscellaneous Information Manual

# NAME

**plakar-query** - query flags shared among many Plakar subcommands

# DESCRIPTION

What follows is a set of command line arguments that many
plakar(1)
subcommands provide to filter snapshots.

The flags may be combined to precisely select snapshots.

The generic filters are as follows.
These can be combined to express specific restrictions.
For example,
'`**-before** 7d **-tag** homelab`'
selects snapshots done in the last seven days tagged with
'homelab'.

**-before** *date*

	Select snapshots older than given
	*date*.
	The date may be in RFC3339 format, as
	"*YYYY*-*mm*-*DD* *HH*:*MM*",
	"*YYYY*-*mm*-*DD* *HH*:*MM*:*SS*",
	"*YYYY*-*mm*-*DD*",
	or
	"*YYYY*/*mm*/*DD*"
	where
	*YYYY*
	is a year,
	*mm*
	a month,
	*DD*
	a day,
	*HH*
	a hour in 24 hour format number,
	*MM*
	minutes and
	*SS*
	the number of seconds.

	Alternatively, human-style intervals like
	"half an hour",
	"a month"
	or
	"2h30m"
	are also accepted.

**-category** *name*

> Select snapshot whose category is
> *name*.

**-environment** *name*

> Select snapshot whose environment is
> *name*.

**-job** *name*

> Select snapshot whose job is
> *name*.

**-latest**

> Select only the latest snapshot.

**-name** *name*

> Select snapshots whose name is
> *name*.

**-perimeter** *name*

> Select snapshots whose perimeter is
> *name*.

**-root** *path*

> Select snapshots whose root directory is
> *path*.
> May be specified multiple time, snapshots are selected if any of the
> given paths matches.

**-since** *date*

> Select snapshots newer than the given
> *date*.
> The accepted format is the same as
> **-before**.

**-tag** *name*

> Select snapshots tagged with
> *name*.
> May be specified multiple times, and multiple tags may be given at the
> same time if comma-separated.

Shorthands to express
**-since**
are also provided:

**-minutes** *n*

**-hours** *n*

**-days** *n*

**-weeks** *n*

**-months** *n*

**-years** *n*

These selects only snapshots done in the last
*n*
days of the week, and only on those days.
When combined, they select a snapshot when either condition is met.
For example,
'`**-mondays** 3 **-thuesdays** 5`'
selects snapshots done in the last three mondays and in the last five
thuesdays.

**-mondays** *n*

**-thuesdays** *n*

**-wednesdays** *n*

**-thursdays** *n*

**-fridays** *n*

**-saturdays** *n*

**-sundays** *n*

It's also possible to select at most
*n*
snapshot per unit of time.
When combined, these also behave like additions and not restrictions,
i.e.
'`**-per-month** 3 **-per-monday** 5`'
yields three snapshots per month and 5 per monday.

**-per-minute** *n*

**-per-hour** *n*

**-per-day** *n*

**-per-week** *n*

**-per-month** *n*

**-per-year** *n*

**-per-monday** *n*

**-per-thuesday** *n*

**-per-wednesday** *n*

**-per-thursday** *n*

**-per-friday** *n*

**-per-saturday** *n*

**-per-sunday** *n*

Plakar - September 10, 2025
