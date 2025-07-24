<script lang="ts">
  import COLORS from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    renderMonthlyInvestmentTimeline,
    renderYearlyIncomeTimeline,
    renderYearlyTimelineOf
  } from "$lib/income";
  import { ajax, formatCurrency, type Legend, type Income, type IncomeYearlyCard, type Tax, } from "$lib/utils";
  import _ from "lodash";
  import { onMount, onDestroy } from "svelte";
  import { dateRange, setAllowedDateRange } from "../../../store";

  let grossIncome = 0;
  let netTax = 0;
  let incomes: Income[];
  let taxes: Tax[];
  let yearlyCards: IncomeYearlyCard[];

  let monthlyInvestmentTimelineLegends: Legend[] = [];
  let yearlyIncomeTimelineLegends: Legend[] = [];
  let yearlyNetIncomeTimelineLegends: Legend[] = [];
  let yearlyNetTaxTimelineLegends: Legend[] = [];

  const destroy = (): void => {
    [
      "d3-income-timeline", 
      "d3-yearly-income-timeline", 
      "d3-yearly-net_income-timeline", 
      "d3-yearly-net_tax-timeline"
    ].map((id) => {
      const el = document.getElementById(id)
      if (el.firstChild) {
        el.firstChild.remove()
      }
    })
  };

  $: {
    if (!_.isEmpty(incomes)) {
      destroy();

      const ii = _.filter(
          incomes,
          (p) => p.date.isSameOrBefore($dateRange.to) && p.date.isSameOrAfter($dateRange.from)
        )
      monthlyInvestmentTimelineLegends = renderMonthlyInvestmentTimeline(ii);
      grossIncome = _.sumBy(ii, (i) => _.sumBy(i.postings, (p) => -p.amount));
    }
    if (!_.isEmpty(yearlyCards)) {

      const yy = _.filter(
        yearlyCards,
        (p) => p.end_date.isSameOrAfter($dateRange.from)
      )

      yearlyIncomeTimelineLegends = renderYearlyIncomeTimeline(yy);
      yearlyNetIncomeTimelineLegends = renderYearlyTimelineOf(
        "Net Income",
        "net_income",
        COLORS.gain,
        yy
      );
      yearlyNetTaxTimelineLegends = renderYearlyTimelineOf(
        "Net Tax",
        "net_tax",
        COLORS.loss,
        yy
      );

    }

    if (!_.isEmpty(taxes)) {
      const tt =  _.filter(
          taxes,
          (p) => p.end_date.isSameOrAfter($dateRange.from)
        )
      netTax = _.sumBy(taxes, (t) => _.sumBy(t.postings, (p) => p.amount));
    }
  }

  onDestroy(async () => {
    destroy();
  });

  onMount(async () => {
    const data = await ajax("/api/income");
    incomes = data.income_timeline
    taxes = data.tax_timeline
    yearlyCards = data.yearly_cards
    setAllowedDateRange(_.map(incomes, (p) => p.date));
  });
</script>

<section class="section tab-income">
  <div class="container">
    <nav class="level">
      <LevelItem title="Gross Income" value={formatCurrency(grossIncome)} color={COLORS.gainText} />
      <LevelItem title="Net Tax" value={formatCurrency(netTax)} color={COLORS.lossText} />
    </nav>
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard legends={monthlyInvestmentTimelineLegends} clazz="ml-4" />
          <svg id="d3-income-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text="Monthly Income Timeline" />
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetTaxTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_tax-timeline" width="100%" />
        </div>
      </div>
    </div>
    <BoxLabel text="Financial Year Income Timeline" />
  </div>
</section>
