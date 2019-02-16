import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { environment } from 'src/environments/environment';
import { Observable } from 'rxjs';
import { catchError, tap, map } from 'rxjs/operators';
import { Result } from './result';

@Injectable({
  providedIn: 'root'
})
export class WeatherService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('WeatherService');
  }

  private logger: Logger;

  getWeather(): Observable<Result<WeatherSummary>> {
    const url = environment.apiURL + `weather`;
    this.logger.log('Retrieving current weather information', url);
    return this.http.get<WeatherSummary>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved weather', res)),
        map(res => Result.new(res)),
        catchError(this.logger.handleError(`getWeather()`, Result.new<WeatherSummary>(null, 'Unable to retrieve weather')))
      );
  }
}

export interface WeatherSummary {
  current: string;
  oneWord: string;
  forecast: string;
}